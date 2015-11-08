// ssmp_listener.go
package common

import (
    "bytes"
    //"fmt"
    "log"
)

const LISTENER_BUF_LEN	= 20

var listenerChan chan []byte
var SidNext uint32
var IdNext	int
var ServerID string

func InitListener(svrid string) {
	listenerChan = make(chan []byte, LISTENER_BUF_LEN)
	SidNext = 16
	IdNext = 1
	ServerID = svrid
}

func RunListener(disp *SsmpDispatch, outputChan chan []byte) {
	for {
		select {
		case pktBytes := <- listenerChan :
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
			session := NewServerSession(IdNext, SidNext, ServerID, magic, outputChan)
			
			disp.Register(magic, session.BufChan)
		}
	}
}
