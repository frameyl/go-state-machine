package common

import "testing"
import "time"
import "fmt"
import "bytes"

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

func TestSsmpDispatch(t *testing.T) {
    var dispMng DispatchMng
    ssmpDisp := NewSsmpDispatch("Ssmp Connector", SSMP_DISP_CLNT)

    dispMng.Add(ssmpDisp)
    dispMng.Start()

    dispMng.Handle(*samplePkt1)
    
    timer1 := time.NewTicker(time.Second)
    for i := 0; i < 10; i++ {
        select {
        case <- timer1.C:
            fmt.Printf("Bypass %v", ssmpDisp.Bypass)
            if ssmpDisp.Bypass == 1 {
                goto EOW1
            }
        }
    }
EOW1:    
    if ssmpDisp.Bypass < 1 {
        t.Errorf("SSMP Dispatch failed to bypass a non-SSMP packet, RX %v, Bypass %v", ssmpDisp.Rx, ssmpDisp.Bypass)
    }
    
    dispMng.Handle(*ssmpPkt1)

    for i := 0; i < 10; i++ {
        select {
        case <- timer1.C:
            fmt.Printf("Discard %v", ssmpDisp.Discard)
            if ssmpDisp.Discard == 1 {
                goto EOW2
            }
        }
    }
EOW2:
    if ssmpDisp.Discard < 1 {
        t.Errorf("SSMP Dispatch failed to bypass a non-SSMP packet, RX %v, Bypass %v", ssmpDisp.Rx, ssmpDisp.Bypass)
    }
    
    chanInput := make(chan bytes.Reader)
    ssmpDisp.Register(0x11223344aabbccdd, chanInput)
    
    dispMng.Handle(*ssmpPkt1)
    var buf bytes.Reader

    for i := 0; i < 10; i++ {
        select {
        case <- timer1.C:
            fmt.Printf("Handeld %v", ssmpDisp.Handled)
        case buf = <- chanInput:
            goto EOW3
        }
    }
EOW3:
    fmt.Printf("Packet Length %v", buf.Len())
    if buf.Len() == 0 {
        t.Errorf("SSMP Dispatch: didn't get the packet")
    }
}


