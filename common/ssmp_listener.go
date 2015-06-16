// ssmp_listener.go
package common

import (
    "bytes"
    //"fmt"
    "log"
)

const LISTENER_BUF_LEN	= 20

var listenerChan chan bytes.Reader
var SidNext uint32
var IdNext	int
var ServerID string

func InitListener(svrid string) {
	listenerChan = make(chan bytes.Reader, LISTENER_BUF_LEN)
	SidNext = 16
	IdNext = 1
	ServerID = svrid
}

func RunListener(disp *SsmpDispatch) {
	for {
		select {
		case pkt := <- listenerChan :
			magic, _ := ReadMagicNum(&pkt)
			if magic == MAGIC_NIL {
				log.Println("Invalide MAGIC received ", magic)
				continue
			}
			
			msgType, _ := ReadMsgType(&pkt)
			if msgType != MSG_HELLO {
				log.Println("Listener received a invalid type of package ", GetMsgNameByType(msgType))
				continue
			}
			
			// Create a new server session
			session := NewServerSession(IdNext, SidNext, ServerID, magic)
			
			disp.Register(magic, session.BufChan)
		}
	}
}
