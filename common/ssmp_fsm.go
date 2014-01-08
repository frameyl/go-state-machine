package common

import (
    "bytes"
    "io"
    "log"
    "time"
//    "fmt"
)

type Fsm struct {
    Id      int  //unique identifer of the FSM
    State   int  //state of the FSM
    Magic   uint64  //Magic Number of FSM
    Svrid   string  //Server ID
    Sid     uint32  //Session ID of FSM
    //packet recv channel, with packet goroutine
    BufChan  chan bytes.Reader
    //Control channel (start, reset, clean), with scheduler goroutine
    CntlChan chan int
}

type FsmCnt struct {
    Tx          int
    TxHello     int
    TxDisc      int
    Rx          int
    RxHello     int
    RxDisc      int
    Retry       int
}

type FsmClientCnt struct {
    FsmCnt
    TxRequest   int
    TxConfirm   int
    RxReply     int
}

type FsmServerCnt struct {
    FsmCnt
    TxReply     int
    RxRequest   int
    RxConfirm   int
}

const (
    FSM_CMD_START   = iota
    FSM_CMD_PAUSE
    FSM_CMD_RESET
    FSM_CMD_CLEAN
)

const (
    FSM_STATE_IDLE  = iota
    FSM_STATE_INIT
    FSM_STATE_REQ
    FSM_STATE_EST
    FSM_STATE_DISC
)

const FSM_RETRY     = 5
const FSM_TIMEOUT   = 3

var FsmWrite io.Writer

type stateFn func(*Fsm) stateFn

func (fsm *Fsm) RunClient() {
    for state := Initial; state != nil; {
        state = state(fsm)
    }

    close(fsm.BufChan)
    close(fsm.CntlChan)
}

func Initial(fsm *Fsm) stateFn {
    fsm.State = FSM_STATE_INIT
    // Send Hello packet periodicly, to Requesting phase once it get a hello from server with server id
    for i:=0; i<FSM_RETRY; {
        buf := new(bytes.Buffer)
        WritePacketHdr(buf, MSG_HELLO, fsm.Magic, "")
        n, err := FsmWrite.Write(buf.Bytes())
        if err != nil || n != LEN_SSMP_HDR {
            log.Printf("FSM#%v Write Hello failed, %s", err)
            i++
            continue
        }

        timer := time.NewTimer(FSM_TIMEOUT * time.Second)
LSEL:
        select {
        case <- timer.C:
            i++
        case pkt := <-fsm.BufChan:
            msgType, _ := ReadMsgType(&pkt)
            if msgType != MSG_HELLO {
                break LSEL
            }

            magic, _ := ReadMagicNum(&pkt)
            if(magic != fsm.Magic) {
                break LSEL
            }

            timer.Stop()
            fsm.Svrid, _ = ReadServerID(&pkt)
            return Requesting
        }
    }

    return nil
}

func Requesting(fsm *Fsm) stateFn {
    // Send Request packet periodicly, to Established phase once it get a reply from server
    return Established
}

func Established(fsm *Fsm) stateFn {
    // Doing nothing until it get a disconnect command
    return Disconnecting
}

func Disconnecting(fsm *Fsm) stateFn {
    // Send disconnect packet periodicly, to Initial state once it get a disconnect from server
    return nil
}


