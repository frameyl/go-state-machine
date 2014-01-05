package common

import (
    "testing"
    "time"
    "bytes"
)

/*
var ssmp1 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'H', 'e', 'l', 'l', 'o', 0, 0, 0,
        0x11, 0x22, 0x33, 0x44, 0xaa, 0xbb, 0xcc, 0xdd,
        'N', 'o', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var ssmpPkt1 *bytes.Reader = bytes.NewReader(ssmp1)

var ssmp2 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'R', 'e', 'p', 'l', 'y', 0, 0, 0,
        0, 0, 0x1a, 0, 0x2b, 0, 0x3c, 0,
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
var ssmpPkt2 *bytes.Reader = bytes.NewReader(ssmp2)

var ssmpE1 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'r', 'e', 'p', 'l', 'y', '!', 0, 0,
        'm', 'a', 'g', '3', 'M', 'A', 'G', '4',
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
var ssmpEPkt1 *bytes.Reader = bytes.NewReader(ssmpE1)

var sample1 []byte=[]byte{'S', 'S', 'M', 'P', 'V', '1'}
var samplePkt1 *bytes.Reader = bytes.NewReader(sample1)

var sample2 []byte=[]byte{'S', 'S', 'M', 'P', 'V', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var samplePkt2 *bytes.Reader = bytes.NewReader(sample2)
*/

var ticker *time.Ticker = time.NewTicker(10 * time.Millisecond)

func TestSsmpDispatchClnt(t *testing.T) {
    var dispMng DispatchMng
    ssmpDisp := NewSsmpDispatch("Ssmp Connector", SSMP_DISP_CLNT)

    dispMng.Add(ssmpDisp)
    dispMng.Start()

    dispMng.Handle(*samplePkt1)
    
    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{1, 0, 1, 0}},
        func() {
            t.Errorf("SSMP Dispatch failed to bypass a non-SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
    dispMng.Handle(*ssmpPkt1)

    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{2, 0, 1, 1}},
        func() {
            t.Errorf("SSMP Dispatch failed to discard a unknown SSMP packet, %s", ssmpDisp.DispatchCnt)
        },
        10,
    )
    
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

    ssmpDisp.Unregister(0x11223344aabbccdd)
    
    dispMng.Handle(*ssmpPkt1)

    WaitforCondition(
        func() bool {return ssmpDisp.DispatchCnt == DispatchCnt{4, 1, 1, 2}},
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


