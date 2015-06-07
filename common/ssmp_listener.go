// ssmp_listener.go
package common

import (
    "bytes"
    //"fmt"
    //"log"
)

const LISTENER_BUF_LEN	= 20

var listenerChan chan bytes.Reader

func InitListener() {
	listenerChan = make(chan bytes.Reader, LISTENER_BUF_LEN)
}

func RunListener() {
	for {
		select {
		case pkt := <- listenerChan :
			magic, _ := ReadMagicNum(&pkt)
			magic = magic		
			
		}
	}
}
