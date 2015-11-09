package common

import (
	"bytes"
	"log"
	"math/rand"
	"time"
	//"fmt"
	"github.com/looplab/fsm"
)

// OutputChan is the channel for sending packets out
//var OutputChan chan []byte

// MagicChan tell outside that the magic number of a session changed
//var MagicChan chan MagicReg

type MagicReg struct {
	BufChan chan []byte // Session
	Magic   uint64      // Magic Number
}

// Session defines a session for SSMP
type Session struct {
	Id    int    //unique identifer, internally used
	Magic uint64 //Magic Number generated by client
	Svrid string //Server ID
	Sid   uint32 //Session ID allocated by Server

	// packet recv channel, with packet goroutine
	BufChan chan []byte
	// packet output channel, shared with other sessions
	OutputChan chan []byte
	// Channel for magic number change, shared with other sessions
	MagicChan chan MagicReg
	// Control channel (start, reset, clean), with scheduler goroutine
	CntlChan chan int
	// Counters
	SessionCnt
	// FSM (looplab style)
	fsm *fsm.FSM
}

type SessionClient struct {
	Session
	retryTimer *PTimer
	deadTimer  *PTimer
	SessionClientCnt
}

type SessionServer struct {
	Session
	deadTimer *PTimer
	SessionServerCnt
}

type SessionCnt struct {
	Tx      int
	TxHello int
	TxDisc  int
	Rx      int
	RxHello int
	RxDisc  int
	Retry   int
}

type SessionClientCnt struct {
	TxRequest int
	TxConfirm int
	RxReply   int
}

type SessionServerCnt struct {
	TxReply   int
	RxRequest int
	RxConfirm int
}

type SessionCmd int

const (
	// Command for FSM
	S_CMD_NOTHING = iota
	S_CMD_START
	S_CMD_STOP
	S_CMD_PAUSE
	S_CMD_CONTINUE
	S_CMD_CLEAN
)

// send*() define a serial of functions to send protocol packets out
//
// sendPacket():	General call
// sendHello():		Send Hello packets
// sendRequest():	Send Request packets (clinet only)
// sendConfirm():	Send Confirm packets (client only)
// sendClose():		Send Close packets (client or server)
// sendReply():		Send Reply packets (server only)
func (s *Session) sendPacket(mtype MsgType) error {
	buf := new(bytes.Buffer)
	WritePacketHdr(buf, mtype, s.Magic, s.Svrid)
	pktLen := LEN_SSMP_HDR

	if mtype == MSG_REPLY || mtype == MSG_CONFIRM || mtype == MSG_CLOSE {
		WriteSessionID(buf, s.Sid)
		pktLen += LEN_SESSION_ID
	}

	s.OutputChan <- buf.Bytes()

	return nil
}

func (s *Session) sendHello() error {
	return s.sendPacket(MSG_HELLO)
}

func (s *Session) sendRequest() error {
	return s.sendPacket(MSG_REQUEST)
}

func (s *Session) sendConfirm() error {
	return s.sendPacket(MSG_CONFIRM)
}

func (s *Session) sendClose() error {
	return s.sendPacket(MSG_CLOSE)
}

func (s *Session) sendReply() error {
	return s.sendPacket(MSG_REPLY)
}

const SESSION_TIMEOUT_RETRY = 2 * time.Second
const SESSION_TIMEOUT_DEAD = 5 * time.Second

// Timer functions:
//		retryTimerOn
//		retryTimerOff
//		deadTimerOn
//		deadTimerOff
func (cs *SessionClient) retryTimerOn() error {
	cs.retryTimer.TimerOn()
	return nil
}

func (cs *SessionClient) retryTimerOff() error {
	cs.retryTimer.TimerOff()
	return nil
}

func (cs *SessionClient) deadTimerOn() error {
	cs.deadTimer.TimerOn()

	return nil
}

func (cs *SessionClient) deadTimerOff() error {
	cs.deadTimer.TimerOff()
	return nil
}

func (ss *SessionServer) deadTimerOn() error {
	ss.deadTimer.TimerOn()
	return nil
}

func (ss *SessionServer) deadTimerOff() error {
	ss.deadTimer.TimerOff()
	return nil
}

// enterState is a callback and will be called once a state transaction happens
func (s *Session) enterState(e *fsm.Event) {
	log.Println("Session", s.Id, "Entering state", s.fsm.Current(), "with Event", e.Event)
}

// connected is a callback and will be called once the session connects
func (s *Session) connected(e *fsm.Event) {
	log.Println("Session", s.Id, "Connected", "with Event", e.Event)
	s.MagicChan <- MagicReg{nil, s.Magic}
	return
}

func (s *Session) disconnected(e *fsm.Event) {
	log.Println("Session", s.Id, "Disconnected", "with Event", e.Event)
	return
}

func (s *Session) clean(e *fsm.Event) {
	log.Println("Session", s.Id, "cleaned", "with Event", e.Event)

	// Clean SrvID, Magic
	s.Svrid = ""
	s.Sid = 0
	s.Magic = 0

	return
}

func (s *Session) Current() string {
	return s.fsm.Current()
}

// retryTimeout is a callback and will be called after a retry timer expired
func (s *SessionClient) retryTimeout(e *fsm.Event) {
	log.Println("Session", s.Id, "retry Timer expired")

	current := s.fsm.Current()
	switch current {
	case "init":
		s.sendHello()

	case "req":
		s.sendRequest()

	case "close":
		s.sendClose()

	default:
		log.Println("Invalide state when retry timer expired.")
		return
	}

	s.SessionCnt.Retry++
	s.retryTimerOn()

	return
}

func (s *SessionClient) deadTimeout(e *fsm.Event) {
	log.Println("Session", s.Id, "dead Timer expired")

	current := s.fsm.Current()
	switch current {
	case "init", "req", "close":
		// Clean SrvID, Magic
		s.Svrid = ""
		s.Sid = 0
		s.Magic = 0

	default:
		log.Println("Invalide state when dead timer expired.")
		return
	}

	return
}

func (s *SessionServer) deadTimeout(e *fsm.Event) {
	log.Println("Session", s.Id, "cleaned", "with Event", e.Event)

	current := s.fsm.Current()
	switch current {
	case "allocated":
		// Clean Magic
		s.Magic = 0

	default:
		log.Println("Invalide state when dead timer expired.")
		return
	}

	return
}

func NewClientSession(id int, outputChan chan []byte, magicChan chan MagicReg) *SessionClient {
	s := &SessionClient{
		Session: Session{
			Id:         id,
			BufChan:    make(chan []byte, 5),
			CntlChan:   make(chan int, 2),
			OutputChan: outputChan,
			MagicChan:  magicChan,
		},
		retryTimer: NewPTimer(SESSION_TIMEOUT_RETRY),
		deadTimer:  NewPTimer(SESSION_TIMEOUT_DEAD),
	}

	s.fsm = fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "start", Src: []string{"idle"}, Dst: "init"},

			{Name: "stop", Src: []string{"init"}, Dst: "close"},
			{Name: "stop", Src: []string{"req", "est"}, Dst: "close"},

			{Name: "hello_received", Src: []string{"init"}, Dst: "req"},

			{Name: "reply_received", Src: []string{"req"}, Dst: "est"},

			{Name: "disconnect_received", Src: []string{"init"}, Dst: "idle"},
			{Name: "disconnect_received", Src: []string{"req"}, Dst: "idle"},
			{Name: "disconnect_received", Src: []string{"est"}, Dst: "idle"},

			{Name: "pause", Src: []string{"init"}, Dst: "init"},
			{Name: "pause", Src: []string{"req"}, Dst: "req"},
			{Name: "pause", Src: []string{"est"}, Dst: "est"},
			{Name: "pause", Src: []string{"close"}, Dst: "close"},
			{Name: "continue", Src: []string{"init"}, Dst: "init"},
			{Name: "continue", Src: []string{"req"}, Dst: "req"},
			{Name: "continue", Src: []string{"est"}, Dst: "est"},
			{Name: "continue", Src: []string{"close"}, Dst: "close"},

			{Name: "clean", Src: []string{"idle", "init", "req", "est", "close"}, Dst: "idle"},

			{Name: "retry_timeout", Src: []string{"init"}, Dst: "init"},
			{Name: "retry_timeout", Src: []string{"req"}, Dst: "req"},
			{Name: "retry_timeout", Src: []string{"est"}, Dst: "est"},
			{Name: "retry_timeout", Src: []string{"close"}, Dst: "close"},

			{Name: "dead_timeout", Src: []string{"init", "req", "est", "close"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { s.enterState(e) },
			"enter_idle":  func(e *fsm.Event) { s.disconnected(e) },
			"enter_init":  func(e *fsm.Event) { s.sendHello(); s.retryTimerOn(); s.deadTimerOn() },
			"enter_req":   func(e *fsm.Event) { s.sendRequest(); s.retryTimerOn(); s.deadTimerOn() },
			"enter_est":   func(e *fsm.Event) { s.sendConfirm(); s.connected(e) },
			"enter_close": func(e *fsm.Event) { s.sendClose(); s.retryTimerOn(); s.deadTimerOn() },

			"leave_init":  func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"leave_req":   func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"leave_close": func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },

			"before_start":          func(e *fsm.Event) { s.genMagicNumber(e) },
			"before_hello_received": func(e *fsm.Event) { s.recvHello(e) },
			"before_reply_received": func(e *fsm.Event) { s.recvReply(e) },

			"after_pause":    func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"after_continue": func(e *fsm.Event) { s.retryTimerOn(); s.deadTimerOn() },

			"after_clean": func(e *fsm.Event) { s.clean(e) },

			"after_retry_timeout": func(e *fsm.Event) { s.retryTimeout(e) },
			"after_dead_timeout":  func(e *fsm.Event) { s.deadTimeout(e) },
		},
	)
	return s
}

func (s *SessionClient) RunClient() {
	for {
		select {
		case <-s.retryTimer.C():
			s.fsm.Event("retry_timeout")
		case <-s.deadTimer.C():
			s.fsm.Event("dead_timeout")
		case cmd := <-s.CntlChan:
			if cmd == S_CMD_START {
				s.fsm.Event("start")
			} else if cmd == S_CMD_STOP {
				s.fsm.Event("stop")
			} else if cmd == S_CMD_PAUSE {
				s.fsm.Event("pause")
			} else if cmd == S_CMD_CONTINUE {
				s.fsm.Event("continue")
			} else if cmd == S_CMD_CLEAN {
				s.fsm.Event("clean")
			} else {
				log.Println("Invalide command received", cmd)
			}
		case pktBytes := <-s.BufChan:
			// got a packet here
			pkt := bytes.NewReader(pktBytes)

			s.Rx++
			msgType, _ := ReadMsgType(pkt)
			if msgType == MSG_HELLO {
				s.fsm.Event("hello_received", pkt)
			} else if msgType == MSG_REPLY {
				s.fsm.Event("reply_received", pkt)
			} else {
				log.Println("Invalid package received, type", msgType)
			}
		}
	}
}

func (s *SessionClient) recvHello(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxHello++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log.Println("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	s.Svrid, _ = ReadServerID(pkt)
	return
}

func (s *SessionClient) recvReply(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxReply++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log.Println("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	svrid, _ := ReadServerID(pkt)
	if svrid != s.Svrid {
		log.Println("Wrong Server ID", s.Svrid)
		e.Cancel()
		return
	}

	s.Sid, _ = ReadSessionID(pkt)
	return
}

func (s *SessionClient) genMagicNumber(e *fsm.Event) {
	s.Magic = uint64(rand.Int63())
	s.MagicChan <- MagicReg{s.BufChan, s.Magic}
}

func NewServerSession(id int, sid uint32, svrid string, magic uint64, outputChan chan []byte) *SessionServer {
	s := &SessionServer{
		Session: Session{
			Id:         id,
			Sid:        sid,
			Svrid:      svrid,
			Magic:      magic,
			BufChan:    make(chan []byte, 2),
			CntlChan:   make(chan int),
			OutputChan: outputChan,
		},
		deadTimer: NewPTimer(SESSION_TIMEOUT_DEAD),
	}

	s.fsm = fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "start", Src: []string{"idle"}, Dst: "listening"},
			{Name: "stop", Src: []string{"listening", "allocated", "est"}, Dst: "idle"},
			{Name: "hello_received", Src: []string{"listening"}, Dst: "listening"},
			{Name: "request_received", Src: []string{"listening"}, Dst: "allocated"},
			{Name: "confirm_received", Src: []string{"allocated"}, Dst: "est"},
			{Name: "close_received", Src: []string{"est"}, Dst: "listening"},

			{Name: "pause", Src: []string{"allocated"}, Dst: "allocated"},
			{Name: "continue", Src: []string{"allocated"}, Dst: "allocated"},

			{Name: "clean", Src: []string{"idle", "listening", "allocated", "est"}, Dst: "idle"},

			{Name: "dead_timeout", Src: []string{"allocated", "est"}, Dst: "listening"},
		},
		fsm.Callbacks{
			"enter_state":     func(e *fsm.Event) { s.enterState(e) },
			"enter_idle":      func(e *fsm.Event) { s.disconnected(e) },
			"enter_listening": func(e *fsm.Event) { /*s.registerSession(e)*/ },
			"enter_allocated": func(e *fsm.Event) { s.sendReply(); s.deadTimerOn() },
			"enter_est":       func(e *fsm.Event) { s.connected(e) },

			"leave_allocated": func(e *fsm.Event) { s.deadTimerOff() },

			"before_hello_received":   func(e *fsm.Event) { s.recvHello(e); s.sendHello() },
			"before_request_received": func(e *fsm.Event) { s.recvRequest(e) },
			"before_confirm_received": func(e *fsm.Event) { s.recvConfirm(e) },
			"before_close_received":   func(e *fsm.Event) { s.recvClose(e) },

			"after_close_received": func(e *fsm.Event) { s.sendClose() },

			"after_stop":     func(e *fsm.Event) {},
			"after_pause":    func(e *fsm.Event) { s.deadTimerOff() },
			"after_continue": func(e *fsm.Event) { s.deadTimerOn() },

			"after_clean": func(e *fsm.Event) { s.clean(e) },

			"after_dead_timeout": func(e *fsm.Event) { s.deadTimeout(e) },
		},
	)
	return s
}

func (s *SessionServer) RunServer() {
	for {
		select {
		case <-s.deadTimer.C():
			s.fsm.Event("dead_timeout")
		case cmd := <-s.CntlChan:
			if cmd == S_CMD_START {
				s.fsm.Event("start")
			} else if cmd == S_CMD_STOP {
				s.fsm.Event("stop")
			} else if cmd == S_CMD_PAUSE {
				s.fsm.Event("pause")
			} else if cmd == S_CMD_CONTINUE {
				s.fsm.Event("continue")
			} else if cmd == S_CMD_CLEAN {
				s.fsm.Event("clean")
			} else {
				log.Println("Invalide command received", cmd)
			}
		case pktBytes := <-s.BufChan:
			// got a packet here
			s.Rx++

			pkt := bytes.NewReader(pktBytes)
			msgType, _ := ReadMsgType(pkt)
			if msgType == MSG_HELLO {
				s.fsm.Event("hello_received", pkt)
			} else if msgType == MSG_REQUEST {
				s.fsm.Event("request_received", pkt)
			} else if msgType == MSG_CONFIRM {
				s.fsm.Event("confirm_received", pkt)
			} else if msgType == MSG_CLOSE {
				s.fsm.Event("disconnect_received", pkt)
			} else {
				log.Println("Invalid package received, type", msgType)
			}
		}
	}
}

func (s *SessionServer) recvHello(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxHello++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log.Println("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	return
}

func (s *SessionServer) recvRequest(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxRequest++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log.Println("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	svrid, _ := ReadServerID(pkt)
	if svrid != s.Svrid {
		log.Println("Wrong Server ID", s.Svrid)
		e.Cancel()
		return
	}

	return
}

func (s *SessionServer) recvConfirm(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxConfirm++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log.Println("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	svrid, _ := ReadServerID(pkt)
	if svrid != s.Svrid {
		log.Println("Wrong Server ID", s.Svrid)
		e.Cancel()
		return
	}

	sid, _ := ReadSessionID(pkt)
	if sid != s.Sid {
		log.Println("Wrong Session ID", s.Sid)
		e.Cancel()
		return
	}

	return
}

func (s *SessionServer) recvClose(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxDisc++
	//magic, _ := ReadMagicNum(pkt)
	//if magic != s.Magic {
	//	log.Println("Wrong Magic number", s.Magic)
	//	e.Cancel()
	//	return
	//}

	svrid, _ := ReadServerID(pkt)
	if svrid != s.Svrid {
		log.Println("Wrong Server ID", s.Svrid)
		e.Cancel()
		return
	}

	sid, _ := ReadSessionID(pkt)
	if sid != s.Sid {
		log.Println("Wrong Session ID", s.Sid)
		e.Cancel()
		return
	}

	return
}
