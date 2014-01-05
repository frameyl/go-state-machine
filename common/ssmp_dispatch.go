package common

import (
    "bytes"
    "fmt"
)

type SsmpDispatch struct {
    // Name of dispatch
    name string
    // Mode of Dispatch
    mode int
    // Channel of packet buffer
    bufChan chan bytes.Reader
    // Channel of control message
    cntlChan chan int
    // Channel of register
    regChan chan SsmpDispatchReg
    // Map from Magic Number to FSM input channel
    mapFsm map[uint64] chan bytes.Reader
    // Default input channel for unmapped packets
    defaultFsm chan bytes.Reader
    // Counters
    DispatchCnt
}

type SsmpDispatchReg struct {
    Magic uint64
    BufChan chan bytes.Reader
}

const SSMP_DISP_CLOSE       = 0
const SSMP_DISP_RESET       = 1

const (
    SSMP_DISP_SVR   = iota
    SSMP_DISP_CLNT
)

func NewSsmpDispatch(dispName string, mode int) *SsmpDispatch {
    disp := &SsmpDispatch{
                dispName,
                mode,
                make(chan bytes.Reader),
                make(chan int),
                make(chan SsmpDispatchReg),
                make(map[uint64] chan bytes.Reader),
                make(chan bytes.Reader),
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
            // Not a SSMP packet
            //fmt.Printf("debug3")
            if isSsmpPkt, _ := IsSsmpPacket(&packet); !isSsmpPkt {
                if nextStep != nil {
                    nextStep <- packet
                }
                disp.Bypass++
                continue
            }

            // With a unknow message name
            if msgType, _ := ReadMsgType(&packet); msgType == MSG_UNKNOWN {
                disp.Discard++
                continue
            }

            // Dispatch packet according to the magic number
            magic, _ := ReadMagicNum(&packet)

            fsmChan, ok := disp.mapFsm[magic]
            if ok {
                fsmChan <- packet
                disp.Handled++
                continue
            } else {
                // Handle it as an incoming session if I'm a server
                // Discard if I'm a client
                if disp.mode == SSMP_DISP_CLNT {
                    disp.Discard++
                    continue
                } else if disp.mode == SSMP_DISP_SVR {
                    disp.defaultFsm <- packet
                    disp.Handled++
                    continue
                }
            }

        case cmd := <-disp.cntlChan:
            // Command 0 means terminate the routine
            if cmd == SSMP_DISP_CLOSE {
                close(disp.bufChan)
                close(disp.cntlChan)
                close(disp.regChan)
                return nil
            } else if cmd == SSMP_DISP_RESET {
                disp.mapFsm = make(map[uint64] chan bytes.Reader)
            }

        case reg := <-disp.regChan:
            if reg.BufChan != nil {
                if _, ok := disp.mapFsm[reg.Magic]; ok {
                    fmt.Printf("Try to register a MagicNum already existed %v\n", reg.Magic)
                    continue
                }
                disp.mapFsm[reg.Magic] = reg.BufChan
                fmt.Printf("Register a MagicNum %X\n", reg.Magic)
            } else {
                if _, ok := disp.mapFsm[reg.Magic]; !ok {
                    fmt.Printf("Try to unregister a MagicNum not existed %v\n", reg.Magic)
                    continue
                }
                delete(disp.mapFsm, reg.Magic)
            }
        }
    }

    return nil
}


