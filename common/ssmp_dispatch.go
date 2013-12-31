package common

import (
    "bytes"
    "fmt"
    "encoding/binary"
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
                make(map[uint64] chan bytes.Reader) }
    
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

func (disp *SsmpDispatch)Handle(nextStep chan bytes.Reader) (err error) {
    for {
        select {
        case packet := <-disp.bufChan:
            if packet.Len() < LEN_SSMP_HDR {
                return fmt.Errorf("SSMP Dispatch get a invalide packet, len = %v", packet.Len())
            }

            magicBytes := make([]byte, 8)
            n, _ := packet.ReadAt(magicBytes, OFF_MAGIC_NUMBER)
            if n != LEN_MAGIC_NUMBER {
                fmt.Printf("Found error during read magic number, len %v", n)
                if nextStep != nil {
                    nextStep <- packet
                }
                continue
            }
            magic := binary.BigEndian.Uint64(magicBytes)

            fsmChan, ok := disp.mapInput[magic]
            if ok {
                fsmChan <- packet
            } else {
                // A new incoming session for server
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
                if _, ok := disp.mapInput[reg.Magic]; ok {
                    fmt.Printf("Try to unregister a MagicNum not existed %v", reg.Magic)
                    continue
                }
                delete(disp.mapInput, reg.Magic)
            }
        }
    }
    
    return nil
}


