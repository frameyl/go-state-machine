package common

import (
    "testing"
    "time"
    "bytes"
    "log"
)

var ticker *time.Ticker = time.NewTicker(10 * time.Millisecond)

func TestSsmpDispatchClnt(t *testing.T) {
    log.SetFlags(log.Ldate | log.Lmicroseconds)

    var dispMng DispatchMng
    ssmpDisp := NewSsmpDispatch("Ssmp Client", SSMP_DISP_CLNT)

    dispMng.Add(ssmpDisp)
    dispMng.Start()

    // Input a short packet which is not a valid SSMP packet
    dispMng.Handle(*samplePkt1)
    
    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{1, 0, 1, 0}},
        func() {
            t.Errorf("SSMP Dispatch failed to bypass a non-SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
    // Input a SSMP hello packet
    dispMng.Handle(*ssmpPkt1)

    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{2, 0, 1, 1}},
        func() {
            t.Errorf("SSMP Dispatch failed to discard a unknown SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
    // Register a FSM channel and input a SSMP hello packet
    chanInput := make(chan bytes.Reader)
    ssmpDisp.Register(0x11223344aabbccdd, chanInput)
    
    dispMng.Handle(*ssmpPkt1)

    WaitforBufChannel(
        chanInput, 
        func(buf bytes.Reader) {
            if buf.Len() != 64 {
                t.Errorf("SSMP Dispatch distributed a error packet, len %v", buf.Len())
            }
        },
        func() {
            t.Errorf("SSMP Dispatch: didn't get the packet")
        }, 
        10,
    )

    // Unregister a FSM channel and input a SSMP hello packet
    ssmpDisp.Unregister(0x11223344aabbccdd)
    
    dispMng.Handle(*ssmpPkt1)

    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{4, 1, 1, 2}},
        func() {
            t.Errorf("SSMP Dispatch failed to discard a SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
    // Register a FSM channel and input a SSMP reply packet
    ssmpDisp.Register(0x1a002b003c00, chanInput)
    dispMng.Handle(*ssmpPkt2)
    
    WaitforBufChannel(
        chanInput, 
        func(buf bytes.Reader) {
            if buf.Len() != 68 {
                t.Errorf("SSMP Dispatch distributed a error packet, len %v", buf.Len())
            }
        },
        func() {
            t.Errorf("SSMP Dispatch: didn't get the packet")
        }, 
        10,
    )
    
    // Reset the dispatch and input a SSMP packet
    ssmpDisp.Reset()
    dispMng.Handle(*ssmpPkt2)
    
    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{6, 2, 1, 3}},
        func() {
            t.Errorf("SSMP Dispatch failed to discard a SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
}

func WaitforCondition(cond func() bool, timeoutHnl func(), timeout int) {
    for i := 0; i < timeout; i++ {
        select {
        case <- ticker.C:
            if cond() {
                return
            }
        }
    }
    
    timeoutHnl()
}

func WaitforBufChannel(bufchan chan bytes.Reader, action func(buf bytes.Reader), timeoutHnl func(), timeout int) {
    for i := 0; i < timeout; {
        select {
        case <- ticker.C:
            i++
        case buf := <- bufchan:
            action(buf)
            return
        }
    }

    timeoutHnl()
}


