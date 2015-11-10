// main
package main

import (
	"fmt"
	"log"
	"go-state-machine/common"
    "time"
    "bytes"
)

var ChanServer2Client chan []byte
var ChanClient2Server chan []byte

var DispMngServer common.DispatchMng
var DispMngClient common.DispatchMng

var SsmpDispServer *common.SsmpDispatch
var SsmpDispClient *common.SsmpDispatch

var SessionGroup *common.SessionGroupClient

func main() {
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

	log.Println("Start Client...")
	start_client()
    
    for {
        SessionGroup.Dump()
        SsmpDispClient.DumpCounters()
        SsmpDispServer.DumpCounters()
        time.Sleep(time.Second)
    }
}

func init_server() {
	// Initialize dispatch for server
    SsmpDispServer = common.NewSsmpDispatch("ServerMi1", common.SSMP_DISP_SVR)
    listener := common.NewSsmpListener("ServerMine1", 256, 1, ChanServer2Client)

    SsmpDispServer.SetListener(listener)
    
    go listener.RunListener(SsmpDispServer)

    DispMngServer.Add(SsmpDispServer)
	
    DispMngServer.Start()
}

func init_client() {
	// Initialize dispatch for client
    SsmpDispClient = common.NewSsmpDispatch("ClientMi1", common.SSMP_DISP_CLNT)

    DispMngClient.Add(SsmpDispClient)
	
	// Initialize session group for client
	SessionGroup = common.NewSessionGroupClient(1, 10, SsmpDispClient, ChanClient2Server)

    DispMngClient.Start()
}

func start_client() {
	SessionGroup.Start()
}

