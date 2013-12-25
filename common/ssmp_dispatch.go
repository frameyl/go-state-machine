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
    cntlChan chan SsmpDispCmd
    // Map from Magic Number to FSM input channel
    mapInput map[uint64] chan bytes.Reader
}

const SSMP_DISP_CLOSE       = 0
const SSMP_DISP_REG         = 1
const SSMP_DISP_UNREG       = 2

type SsmpDispCmd struct {
    Cmd         int
    MagicNum    uint64
    FsmChan     chan bytes.Reader
}

func NewSsmpDispatch(n string) *SsmpDispatch {
    disp := &SsmpDispatch{name: n}
    
    return disp
}

func (disp SsmpDispatch)Name() string {
    return disp.name
}

func (disp SsmpDispatch)GetBufChan() chan bytes.Reader {
    return disp.bufChan
}

func (disp SsmpDispatch)GetCntlChan() chan SsmpDispCmd {
    return disp.cntlChan
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
            if cmd.Cmd == SSMP_DISP_CLOSE {
                break
            } else if cmd.Cmd == SSMP_DISP_REG {
                if _, ok := disp.mapInput[cmd.MagicNum]; ok {
                    fmt.Printf("Try to register a MagicNum already existed %v", cmd.MagicNum)
                    continue
                }
                disp.mapInput[cmd.MagicNum] = cmd.FsmChan
            } else if cmd.Cmd == SSMP_DISP_UNREG {
                if _, ok := disp.mapInput[cmd.MagicNum]; ok {
                    fmt.Printf("Try to unregister a MagicNum not existed %v", cmd.MagicNum)
                    continue
                }
                delete(disp.mapInput, cmd.MagicNum)
            }
        }
    }
    
    return nil
}


