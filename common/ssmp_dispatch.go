package common

import (
    "bytes"
    "fmt"
)

type SsmpDispatch struct {
    // Name of dispatch
    name string
    // Channel of packet buffer
    bufChan chan bytes.Reader
    // Channel of control message
    cntlChan chan int
    // Channel of register
    regChan chan SsmpDispatchReg
    // Map from Magic Number to FSM input channel
    mapInput map[uint64] chan bytes.Reader
    // Counters
    DispatchCnt
}

type SsmpDispatchReg struct {
    Magic uint64
    BufChan chan bytes.Reader
}

const SSMP_DISP_CLOSE       = 0
const SSMP_DISP_RESET       = 1

func NewSsmpDispatch(dispName string) *SsmpDispatch {
    disp := &SsmpDispatch{
                dispName,
                make(chan bytes.Reader),
                make(chan int),
                make(chan SsmpDispatchReg),
                make(map[uint64] chan bytes.Reader),
                DispatchCnt{} }

    return disp
}

func (disp SsmpDispatch)Name() string {
    return disp.name
}

func (disp SsmpDispatch)GetBufChan() chan bytes.Reader {
    return disp.bufChan
}

func (disp SsmpDispatch)Close() error {
    disp.cntlChan <- SSMP_DISP_CLOSE
    return nil
}

func (disp SsmpDispatch)Reset() error {
    disp.cntlChan <- SSMP_DISP_RESET
    return nil
}

func (disp SsmpDispatch)Register(magic uint64, bufChan chan bytes.Reader) error {
    disp.regChan <- SsmpDispatchReg{magic, bufChan}
    return nil
}

func (disp SsmpDispatch)Unregister(magic uint64) error {
    disp.regChan <- SsmpDispatchReg{magic, nil}
    return nil
}

func (disp SsmpDispatch)GetCnt() DispatchCnt {
    return disp.DispatchCnt
}

func (disp *SsmpDispatch)Handle(nextStep chan bytes.Reader) (err error) {
    for {
        select {
        case packet := <-disp.bufChan:
            disp.Rx++
            // Packet length is less than the SSMP header
            if packet.Len() < LEN_SSMP_HDR {
                // Not a valid SSMP packet, bypass it
                nextStep <- packet
                disp.Bypass++
            }

            // Not a SSMP packet
            if isSsmpPkt, _ := IsSsmpPacket(&packet); !isSsmpPkt {
                nextStep <- packet
                disp.Bypass++
            }

            // With a unknow message name
            if msgType, _ := ReadMsgType(&packet); msgType == MSG_UNKNOWN {
                disp.Discard++
            }

            magic, _ := ReadMagicNum(&packet)

            fsmChan, ok := disp.mapInput[magic]
            if ok {
                fsmChan <- packet
                disp.Handled++
            } else {
                // Discard if I'm a client
                // Handle it as an incoming session if I'm a server
                // FIXME
            }

        case cmd := <-disp.cntlChan:
            // Command 0 means terminate the routine
            if cmd == SSMP_DISP_CLOSE {
                close(disp.bufChan)
                close(disp.cntlChan)
                close(disp.regChan)
                return nil
            } else if cmd == SSMP_DISP_RESET {
                disp.mapInput = make(map[uint64] chan bytes.Reader)
            }

        case reg := <-disp.regChan:
            if reg.BufChan != nil {
                if _, ok := disp.mapInput[reg.Magic]; ok {
                    fmt.Printf("Try to register a MagicNum already existed %v", reg.Magic)
                    continue
                }
                disp.mapInput[reg.Magic] = reg.BufChan
            } else {
                if _, ok := disp.mapInput[reg.Magic]; !ok {
                    fmt.Printf("Try to unregister a MagicNum not existed %v", reg.Magic)
                    continue
                }
                delete(disp.mapInput, reg.Magic)
            }
        }
    }

    return nil
}


