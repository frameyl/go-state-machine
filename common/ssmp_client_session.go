package common

import (
	"bytes"
	//"log"
	"math/rand"
	//"time"
	//"fmt"
	"github.com/frameyl/log4go"
	"github.com/looplab/fsm"
)

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

// retryTimeout is a callback and will be called after a retry timer expired
func (s *SessionClient) retryTimeout(e *fsm.Event) {
	log4go.Fine("Session", s.Id, "retry Timer expired")

	current := s.fsm.Current()
	switch current {
	case "init":
		s.sendHello()

	case "req":
		s.sendRequest()

	case "close":
		s.sendClose()

	default:
		log4go.Error("Invalid state when retry timer expired.", current)
		return
	}

	s.SessionCnt.Retry++
	s.retryTimerOn()

	return
}

func (s *SessionClient) deadTimeout(e *fsm.Event) {
	log4go.Fine("Session", s.Id, "dead Timer expired")

	current := s.fsm.Current()
	switch current {
	case "init", "req", "close":
		// Clean SrvID, Magic
		s.Svrid = ""
		s.Sid = 0
		s.Magic = 0

	default:
		log4go.Error("Invalid state when dead timer expired.", current)
		return
	}

	return
}

func (s *SessionClient) recvHello(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxHello++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log4go.Trace("Wrong Magic number", s.Magic)
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
		log4go.Trace("Wrong Magic number", s.Magic)
		e.Cancel()
		return
	}

	svrid, _ := ReadServerID(pkt)
	if svrid != s.Svrid {
		log4go.Trace("Wrong Server ID", s.Svrid)
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
				log4go.Error("Invalide command received", cmd)
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
				log4go.Error("Invalid package received, type", msgType)
			}
		}
	}
}
