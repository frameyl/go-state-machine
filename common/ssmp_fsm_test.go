package common

import (
    "testing"
    "bytes"
    //"fmt"
)

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


