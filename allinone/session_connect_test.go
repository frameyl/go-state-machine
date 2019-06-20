package main

import (
    //"fmt"
    //"log"
    "protocols-emu/ssmp"
    "time"
    "bytes"
    "testing"
    "github.com/frameyl/log4go"
)

var tsize = 1000

func BenchmarkSessionsConnecting(b *testing.B) {
    log4go.LoadConfiguration("log.xml")

    ChanClient2Server = make(chan []byte, 100)
    ChanServer2Client = make(chan []byte, 100)

    log4go.Info("Bench Init Client...")
    init_client(tsize)
    log4go.Info("Bench Init Server...")
    init_server()

    go func() {
        for {
            packet := <- ChanClient2Server
            DispMngServer.Handle(packet)
            
            pkt := bytes.NewReader(packet)
            log4go.Fine("Receive a packet from client", common.DumpSsmpPacket(pkt))
        }
    }()

    go func() {
        for {
            packet := <- ChanServer2Client
            DispMngClient.Handle(packet)
            
            pkt := bytes.NewReader(packet)
            log4go.Fine("Receive a packet from server", common.DumpSsmpPacket(pkt))
        }
    }()

    b.ResetTimer()
    
    log4go.Info("Bench Start Client...")
    start_client()
    
    for {
        SessionGroup.Stats()
        //log4go.Info(SsmpDispClient.DumpCounters())
        //log4go.Info(SsmpDispServer.DumpCounters())
        time.Sleep(10*time.Millisecond)
        
        if SessionGroup.Established == tsize {
            break
        }
    }
}
