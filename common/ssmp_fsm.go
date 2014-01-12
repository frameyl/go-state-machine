package common

import (
    "bytes"
    "io"
    "log"
    "time"
    "fmt"
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

type FsmCmd int

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

func (fsm *Fsm) SendPacket(mtype MsgType) error {
    buf := new(bytes.Buffer)
    WritePacketHdr(buf, mtype, fsm.Magic, fsm.Svrid)
    pktLen := LEN_SSMP_HDR

    if mtype == MSG_REPLY || mtype == MSG_CONFIRM || mtype == MSG_CLOSE {
        WriteSessionID(buf, fsm.Sid)
        pktLen += LEN_SESSION_ID
    }

    n, err := FsmWrite.Write(buf.Bytes())
    if err != nil {
        return fmt.Errorf("FSM#%v Write Packet failed, error: %s", err)
    } else if n != pktLen {
        return fmt.Errorf("FSM#%v Write Packet failed, exp len %v, actual %v", pktLen, n)
    }

    return nil
}

func (fsm *Fsm) WaitForPacket(mtype MsgType, timer *time.Timer) (bool, *bytes.Reader, FsmCmd) {
    select {
    case <- timer.C:
        return false, nil, FSM_CMD_TIMEOUT
    case pkt := <-fsm.BufChan:
        msgType, _ := ReadMsgType(&pkt)
        if msgType != mtype {
            return false, &pkt, FSM_CMD_WAIT
        }
        return true, &pkt, FSM_CMD_NOTHING
    case cmd := <-fsm.CntlChan:
        if cmd == FSM_CMD_PAUSE {
            timer.Stop()
        } else if cmd == FSM_CMD_RESET {
            timer = time.NewTimer(FSM_TIMEOUT * time.Second)
        }
        return false, nil, FsmCmd(cmd)
    }
    return false, nil, FSM_CMD_NOTHING
}

/*
func (fsm *Fsm) WaitForPacket(isExpected func(*bytes.Reader) bool, action func(*bytes.Reader) stateFn) (stateFn, error) {
    timer := time.NewTimer(FSM_TIMEOUT * time.Second)
    defer timer.Stop()
LSEL:
    select {
    case <- timer.C:
        return nil, nil
    case pkt := <-fsm.BufChan:
        if isExpected(&pkt) {
            return action(&pkt), nil
        } else {
            break LSEL
        }
    case cmd := <-fsm.CntlChan:
        if cmd == FSM_CMD_PAUSE {
            timer.Stop()
            break LSEL
        } else if cmd == FSM_CMD_RESET {
            timer = time.NewTimer(FSM_TIMEOUT * time.Second)
            break LSEL
        } else if cmd == FSM_CMD_CLEAN {
            return nil, fmt.Errorf("CLEAN")
        }
    }
    return nil, fmt.Errorf("Unexpected Error")
}
*/

func Initial(fsm *Fsm) stateFn {
    fsm.State = FSM_STATE_INIT
    // Send Hello packet periodicly, to Requesting phase once it get a hello from server with server id
    for i:=0; i<FSM_RETRY; {
        if err := fsm.SendPacket(MSG_HELLO); err != nil {
            i++
            log.Printf("%s", err)
            continue
        }

        timer := time.NewTimer(FSM_TIMEOUT * time.Second)
        defer timer.Stop()

        for {
            got, pkt, cmd := fsm.WaitForPacket(MSG_HELLO, timer)
            if got {
                magic, _ := ReadMagicNum(pkt)
                if(magic != fsm.Magic) {
                    continue
                }
                fsm.Svrid, _ = ReadServerID(pkt)
                return Requesting
            }
            if cmd == FSM_CMD_WAIT {
                continue
            } else if cmd == FSM_CMD_CLEAN {
                log.Printf("FSM#%v Got Clean command", fsm.Id)
                return nil
            } else {
                i++
                break
            }
        }
    }

    log.Printf("FSM#%v Initial Failed", fsm.Id)
    return nil
}

func Requesting(fsm *Fsm) stateFn {
    // Send Request packet periodicly, to Established phase once it get a reply from server
    return Established
}

func Established(fsm *Fsm) stateFn {
    // Doing nothing until it get a disconnect command
    return Closing
}

func Closing(fsm *Fsm) stateFn {
    // Send disconnect packet periodicly, to Initial state once it get a disconnect from server
    return nil
}


