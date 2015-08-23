package common

import (
    "testing"
    "bytes"
//    "fmt"
    "log"
)

func TestSessionClient(t *testing.T) {
	sclient := NewClientSession(1)
	
	OutputChan = make(chan []byte)
	
    // Run Client
	go sclient.RunClient()
	
    // Strat Client
	sclient.CntlChan <- S_CMD_START
    
    // Wait for a hello packet
    pktBytes := <- OutputChan
    
    pkt := bytes.NewReader(pktBytes)
    log.Println("Got a packet\n", DumpSsmpPacket(pkt))
    
    if isSsmpPkt, _ := IsSsmpPacket(pkt); !isSsmpPkt {
        t.Errorf("Got a invalide packet")
    }
    
    // Get Magic Number from client
    magic, _ := ReadMagicNum(pkt)
    
    // Input a Hello response to client
    buf := new(bytes.Buffer)
    WritePacketHdr(buf, MSG_HELLO, magic, "Server1")
 
    sclient.BufChan <- buf.Bytes()
    
    // Wait for Request
    pktBytes = <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))
    
    if isSsmpPkt, _ := IsSsmpPacket(pkt); !isSsmpPkt {
        t.Errorf("Got a invalide packet")
    }
    
   // Input a Reply to client
    buf = new(bytes.Buffer)
    WritePacketHdr(buf, MSG_REPLY, magic, "Server1")
    WriteSessionID(buf, 0xFF)
 
    sclient.BufChan <- buf.Bytes()
    
    // Wait for Confirm
    pktBytes = <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))

    if isSsmpPkt, _ := IsSsmpPacket(pkt); !isSsmpPkt {
        t.Errorf("Got a invalide packet")
    }
    
    if sclient.Current() != "est" {
        t.Errorf("Client state machine didn't enter est state!")
    }
}

func TestSessionServer(t *testing.T) {
    magic := uint64(0xABCDABCDEFEF)
	sserver := NewServerSession(1, 999, "Test1", magic)
	
	OutputChan = make(chan []byte)
	
    // Run Server
	go sserver.RunServer()
	
    // Strat Client
	sserver.CntlChan <- S_CMD_START
    
    // Input a hello packet
    buf := new(bytes.Buffer)
    WritePacketHdr(buf, MSG_HELLO, magic, "")
    
    sserver.BufChan <- buf.Bytes()
    
    // Wait for a hello packet
    pktBytes := <- OutputChan
    
    log.Println("Got a packet\n", DumpSsmpPacket(bytes.NewReader(pktBytes)))
    
    // Input a request packet
    buf = new(bytes.Buffer)
    WritePacketHdr(buf, MSG_REQUEST, magic, "Test1")
    
    sserver.BufChan <- buf.Bytes()
    
    // Wait for a reply
    pktBytes = <- OutputChan
    
    pkt := bytes.NewReader(pktBytes)
    log.Println("Got a packet\n", DumpSsmpPacket(pkt))
    sid, _ := ReadSessionID(pkt)
    
    // Input a confirm
    buf = new(bytes.Buffer)
    WritePacketHdr(buf, MSG_CONFIRM, magic, "Test1")
    WriteSessionID(buf, sid)
    
    sserver.BufChan <- buf.Bytes()

    WaitforCondition(
        func() bool {return sserver.Current() == "est"},
        func() {
            t.Errorf("Server state machine didn't enter est state! %s", sserver.Current())
        },
        10,
    )    
}



