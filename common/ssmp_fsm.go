package common

import (
	"bytes"
	//"io"
	"log"
	"time"
	//"fmt"
	"github.com/looplab/fsm"
)

type Session struct {
	Id    int    //unique identifer of the FSM
	Magic uint64 //Magic Number of FSM
	Svrid string //Server ID
	Sid   uint32 //Session ID of FSM
	// packet recv channel, with packet goroutine
	BufChan chan bytes.Reader
	// Control channel (start, reset, clean), with scheduler goroutine
	CntlChan chan int
	// Counters
	SessionCnt
	// FSM (looplab style)
	fsm *fsm.FSM
}

type SessionClient struct {
	Session
	retryTimer *time.Timer
	deadTimer *time.Timer
	SessionClientCnt
}

func NewClientSession(id int) *Session {
	s := &Session{
		Id: id,
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
			{Name: "retry_timeout", Src: []string{"close"}, Dst: "close},

			{Name: "dead_timeout", Src: []string{"init", "req", "est", "close"}, Dst: "idle"},			
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { s.enterState(e) },
			"enter_idle": func(e *fsm.Event) { s.disconnected(e) },
			"enter_init": func(e *fsm.Event) { s.sendHello(); s.retryTimerOn(); s.deadTimerOn() },
			"enter_req": func(e *fsm.Event) { s.sendRequest(); s.retryTimerOn(); s.deadTimerOn() },
			"enter_est": func(e *fsm.Event) { s.established(e) },
			"enter_close": func(e *fsm.Event) { s.sendClose(); s.retryTimerOn(); s.deadTimerOn() },

			"leave_init": func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"leave_req": func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"leave_close": func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			
			"after_pause": func(e *fsm.Event) { s.retryTimerOff(); s.deadTimerOff() },
			"after_continue": func(e *fsm.Event) { s.retryTimerOn(); s.deadTimerOn() },
			
			"after_clean": func(e *fsm.Event) { s.clean(e) },
			
			"after_retry_timeout": func(e *fsm.Event) { s.retryTimeout(e) },
			"after_dead_timeout": func(e *fsm.Event) { s.deadTimeout(e) },
		},
	)
	return s
}

func NewServerSession(id int) *Session {
	s := &Session{
		Id: id,
	}

	s.fsm = fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "start", Src: []string{"idle"}, Dst: "init"},
			{Name: "stop", Src: []string{"init", "allocated", "est"}, Dst: "close"},
			{Name: "request_received", Src: []string{"init"}, Dst: "allocated"},
			{Name: "confirm_received", Src: []string{"allocated"}, Dst: "est"},
			{Name: "disconnect_received", Src: []string{"est"}, Dst: "close"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) { d.enterState(e) },
		},
	)
	return s
}


type SessionServer struct {
	Session
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
	FSM_CMD_NOTHING = iota
	FSM_CMD_START
	FSM_CMD_PAUSE
	FSM_CMD_RESET
	FSM_CMD_CLEAN
	// The following is internal command
	FSM_CMD_TIMEOUT
	FSM_CMD_WAIT
	FSM_CMD_RETRY
)

const (
	FSM_STATE_IDLE = iota
	FSM_STATE_INIT
	FSM_STATE_REQ
	FSM_STATE_EST
	FSM_STATE_CLOSE
)

const FSM_RETRY = 5
const FSM_TIMEOUT = 3

var FsmWrite chan bytes.Buffer

type stateFn func(*Fsm) stateFn

func (fsm *Fsm) RunClient() {
	for state := Initial; state != nil; {
		state = state(fsm)
	}

	close(fsm.BufChan)
	close(fsm.CntlChan)
}

func (fsm *Fsm) SendPacket(mtype MsgType) error {
	buf := new(bytes.Buffer)
	WritePacketHdr(buf, mtype, fsm.Magic, fsm.Svrid)
	pktLen := LEN_SSMP_HDR

	if mtype == MSG_REPLY || mtype == MSG_CONFIRM || mtype == MSG_CLOSE {
		WriteSessionID(buf, fsm.Sid)
		pktLen += LEN_SESSION_ID
	}

	FsmWrite <- *buf

	return nil
}

func (fsm *Fsm) WaitForPacket(mtype MsgType, timer *time.Timer) (*bytes.Reader, FsmCmd) {
	select {
	case <-timer.C:
		return nil, FSM_CMD_TIMEOUT
	case pkt := <-fsm.BufChan:
		fsm.Rx++
		msgType, _ := ReadMsgType(&pkt)
		if msgType != mtype {
			return &pkt, FSM_CMD_WAIT
		}
		return &pkt, FSM_CMD_NOTHING
	case cmd := <-fsm.CntlChan:
		if cmd == FSM_CMD_PAUSE {
			timer.Stop()
		} else if cmd == FSM_CMD_RESET {
			timer = time.NewTimer(FSM_TIMEOUT * time.Second)
		}
		return nil, FsmCmd(cmd)
	}
	return nil, FSM_CMD_NOTHING
}

func Initial(fsm *Fsm) stateFn {
	fsm.State = FSM_STATE_INIT
	// Send Hello packet periodicly, to Requesting phase once it get a hello from server with server id
	for i := 0; i < FSM_RETRY; {
		if err := fsm.SendPacket(MSG_HELLO); err != nil {
			i++
			log.Printf("%s", err)
			continue
		}

		fsm.Tx++
		fsm.TxHello++
		timer := time.NewTimer(FSM_TIMEOUT * time.Second)
		defer timer.Stop()

		for {
			pkt, cmd := fsm.WaitForPacket(MSG_HELLO, timer)
			if pkt != nil && cmd == FSM_CMD_NOTHING {
				fsm.RxHello++
				magic, _ := ReadMagicNum(pkt)
				if magic != fsm.Magic {
					continue
				}
				fsm.Svrid, _ = ReadServerID(pkt)
				fsm.NextState = FSM_STATE_REQ
				return Requesting
			}
			if cmd == FSM_CMD_WAIT {
				continue
			} else if cmd == FSM_CMD_CLEAN {
				log.Printf("FSM#%v Got Clean command", fsm.Id)
				fsm.NextState = FSM_STATE_IDLE
				return nil
			} else {
				i++
				break
			}
		}

		fsm.Retry++
	}

	log.Printf("FSM#%v Initial Failed", fsm.Id)
	fsm.NextState = FSM_STATE_IDLE
	return nil
}

func Requesting(fsm *Fsm) stateFn {
	// Send Request packet periodicly, to Established phase once it get a reply from server
	fsm.State = FSM_STATE_REQ
	return Established
}

func Established(fsm *Fsm) stateFn {
	// Doing nothing until it get a disconnect command
	fsm.State = FSM_STATE_EST
	return Closing
}

func Closing(fsm *Fsm) stateFn {
	// Send disconnect packet periodicly, to Initial state once it get a disconnect from server
	fsm.State = FSM_STATE_CLOSE
	return nil
}
