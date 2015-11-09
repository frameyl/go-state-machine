// main
package main

import (
	"fmt"
	"log"
	"go-state-machine/common"
)

var ChanServer2Client chan []byte
var ChanClient2Server chan []byte

var DispMngServer common.DispatchMng
var DispMngClient common.DispatchMng

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
		}
	}()

	go func() {
		for {
			packet := <- ChanServer2Client
			DispMngClient.Handle(packet)
		}
	}()

	log.Println("Start Client...")
	start_client()
}

func init_server() {
	// Initialize dispatch for server
    ssmpDispServer := common.NewSsmpDispatch("ServerMi1", common.SSMP_DISP_SVR)

    DispMngServer.Add(ssmpDispServer)
	
    DispMngServer.Start()
}

func init_client() {
	// Initialize dispatch for client
    ssmpDispClient := common.NewSsmpDispatch("ClientMi1", common.SSMP_DISP_CLNT)

    DispMngClient.Add(ssmpDispClient)
	
	// Initialize session group for client
	SessionGroup = common.NewSessionGroupClient(1, 10, ssmpDispClient, ChanClient2Server)

    DispMngClient.Start()
}

func start_client() {
	SessionGroup.Start()
}

