package main

import (
	"fmt"
	"log"
	"go-state-machine/common"
    "time"
    "bytes"
	"testing"
)

func BenchmarkSessionsConnecting(b *testing.B) {
	fmt.Println("Hello World!")
	
    log.SetFlags(log.Ldate | log.Lmicroseconds)

	ChanClient2Server = make(chan []byte, 100)
	ChanServer2Client = make(chan []byte, 100)

	log.Println("Init Client...")
	init_client()
	log.Println("Init Server...")
	init_server()

	go func() {
		for {
			packet := <- ChanClient2Server
			DispMngServer.Handle(packet)
            
            pkt := bytes.NewReader(packet)
            log.Println("Receive a packet from client", common.DumpSsmpPacket(pkt))
		}
	}()

	go func() {
		for {
			packet := <- ChanServer2Client
			DispMngClient.Handle(packet)
            
            pkt := bytes.NewReader(packet)
            log.Println("Receive a packet from server", common.DumpSsmpPacket(pkt))
		}
	}()

	b.ResetTimer()

	log.Println("Start Client...")
	start_client()
    
    for {
        SessionGroup.Dump()
        SsmpDispClient.DumpCounters()
        SsmpDispServer.DumpCounters()
        time.Sleep(time.Second)
		
		if SessionGroup.Established == 1000 {
			break
		}
    }
}
