package common

import (
    "testing"
    "bytes"
//    "fmt"
    "log"
)

func TestSession(t *testing.T) {
	sclient := NewClientSession(1)
	
	OutputChan = make(chan []byte)
	
	go sclient.RunClient()
	
	sclient.CntlChan <- S_CMD_START
    
    pktBytes := <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))
    
    // Input a Hello response to client
    buf := new(bytes.Buffer)
    WritePacketHdr(buf, MSG_HELLO, 0x0, "Server1")
 
    sclient.BufChan <- buf.Bytes()
    
    // Wait for Request
    pktBytes = <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))
    
    // Input a Reply to client
    buf = new(bytes.Buffer)
    WritePacketHdr(buf, MSG_REPLY, 0x0, "Server1")
    WriteSessionID(buf, 0xFF)
 
    sclient.BufChan <- buf.Bytes()
    
    // Wait for Confirm
    pktBytes = <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))
}

/*
func TestFsmInit(t *testing.T) {
    fsmClient := FsmClient{
            Fsm{15, FSM_STATE_IDLE,
                0x11223344, "", 0,
                make(chan bytes.Reader),
                make(chan int),
                FsmCnt{},
                FSM_STATE_IDLE },
            FsmClientCnt{},
    }

    var state stateFn=nil

    FsmWrite = make(chan bytes.Buffer, 1000)

    go func() {
        state = Initial(&fsmClient.Fsm)
    }()

    recv := <-FsmWrite
    reader := bytes.NewReader(recv.Bytes())
    mType, _ := ReadMsgType(reader)
    magic, _ := ReadMagicNum(reader)
    if mType != MSG_HELLO || magic != 0x11223344 {
        t.Errorf("Get wrong hello packet %s %X", GetMsgNameByType(mType), magic)
    }

    buf := new(bytes.Buffer)
    WritePacketHdr(buf, MSG_HELLO, magic, "NERV")
    reader = bytes.NewReader(buf.Bytes())
    fsmClient.BufChan <- *reader

    WaitforCondition(
        func() bool {
            return fsmClient.NextState == FSM_STATE_REQ
        },
        func() {
            t.Errorf("FSM didn't go to request state in 1 second")
        },
        100,
    )
}

func TestFsmInitRetry(t *testing.T) {
    fsmClient := FsmClient{
            Fsm{15, FSM_STATE_IDLE,
                0x11223344, "", 0,
                make(chan bytes.Reader),
                make(chan int),
                FsmCnt{},
                FSM_STATE_IDLE },
            FsmClientCnt{},
    }

    var state stateFn=nil

    FsmWrite = make(chan bytes.Buffer, 1000)

    go func() {
        state = Initial(&fsmClient.Fsm)
    }()

    recv := <-FsmWrite
    reader := bytes.NewReader(recv.Bytes())
    mType, _ := ReadMsgType(reader)
    magic, _ := ReadMagicNum(reader)
    if mType != MSG_HELLO || magic != 0x11223344 {
        t.Errorf("Get wrong hello packet %s %X", GetMsgNameByType(mType), magic)
    }

    // Wait till timeout
    WaitforCondition(
        func() bool {
            return (fsmClient.Retry == 1)
        },
        func() {
            t.Errorf("FSM didn't go to request state in 1 second")
        },
        400,
    )
}
*/

