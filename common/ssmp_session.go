package common

import (
	"bytes"
	"log"
	//"math/rand"
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



// enterState is a callback and will be called once a state transaction happens
func (s *Session) enterState(e *fsm.Event) {
	log.Printf("Session %d (Sid %d) Entering state %s with Event %s", s.Id, s.Sid, s.Current(), e.Event)
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

