package common

import (
	"bytes"
	"log"
	//"math/rand"
	//"time"
	//"fmt"
	"github.com/frameyl/fsm"
	"github.com/frameyl/log4go"
)

func (ss *SessionServer) deadTimerOn() error {
	ss.deadTimer.TimerOn()
	return nil
}

func (ss *SessionServer) deadTimerOff() error {
	ss.deadTimer.TimerOff()
	return nil
}

func (s *SessionServer) deadTimeout(e *fsm.Event) {
	log4go.Fine("Session", s.Id, "cleaned", "with Event", e.Event)

	current := s.fsm.Current()
	switch current {
	case "allocated":
		// Clean Magic
		s.Magic = 0

	default:
		log4go.Error("Invalid state when dead timer expired.", current)
		return
	}

	return
}

func (s *SessionServer) recvHello(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxHello++
	magic, _ := ReadMagicNum(pkt)
	if magic != s.Magic {
		log4go.Trace("Wrong Magic number", s.Magic)
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

	return
}

func (s *SessionServer) recvConfirm(e *fsm.Event) {
	var pkt *bytes.Reader
	pkt = e.Args[0].(*bytes.Reader)

	s.RxConfirm++
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

	sid, _ := ReadSessionID(pkt)
	if sid != s.Sid {
		log4go.Trace("Wrong Session ID", s.Sid)
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
		log4go.Trace("Wrong Server ID", s.Svrid)
		e.Cancel()
		return
	}

	sid, _ := ReadSessionID(pkt)
	if sid != s.Sid {
		log4go.Trace("Wrong Session ID", s.Sid)
		e.Cancel()
		return
	}

	return
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
