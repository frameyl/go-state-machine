// ssmp_listener.go
package common

import (
	"bytes"
	//"fmt"
	"log"
)

const LISTENER_BUF_LEN = 100

type SsmpListener struct {
    ListenerChan chan []byte
    SidNext uint32
    IdNext int
    ServerID string
    OutputChan chan []byte
}

func NewSsmpListener(svrid string, startId int, startSid uint32, outputChan chan []byte) *SsmpListener {
    l := &SsmpListener{
	    ListenerChan: make(chan []byte, LISTENER_BUF_LEN),
	    SidNext: startSid,
	    IdNext: startId,
	    ServerID: svrid,
        OutputChan: outputChan,
    }
    
    return l
}

func (listener *SsmpListener) RunListener(disp *SsmpDispatch) {
	for {
		select {
		case pktBytes := <-listener.ListenerChan:
			pkt := bytes.NewReader(pktBytes)
			magic, _ := ReadMagicNum(pkt)
			if magic == MAGIC_NIL {
				log.Println("Invalide MAGIC received ", magic)
				continue
			}

			msgType, _ := ReadMsgType(pkt)
			if msgType != MSG_HELLO {
				log.Println("Listener received a invalid type of package ", GetMsgNameByType(msgType))
				continue
			}

			// TODO Find a session in pool

			// Create a new server session
			session := NewServerSession(listener.IdNext, listener.SidNext, listener.ServerID, magic, listener.OutputChan)
            go session.RunServer()
			session.CntlChan <- S_CMD_START
			
			log.Printf("%s: a new session was created with ID %d, SID %d, Magic 0x%X.",
					listener.ServerID, listener.IdNext, listener.SidNext, magic)
					
			listener.IdNext++
			listener.SidNext++

			disp.Register(magic, session.BufChan)
		}
	}
}
