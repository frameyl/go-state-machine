package ssmp

//import "fmt"

type Packet_header struct {
    Proto_name      string
    Msg_name        string
    Magic_number    [8]byte
    Server_id       string
}

type Packet_hello struct {
    Packet_header
}

type Packet_request struct {
    Packet_header
}

type Packet_reply struct {
    Packet_header
    Session_id      uint32
}

type Packet_confirm struct {
    Packet_header
    Session_id      uint32
}

func (hdr *Packet_header)encode(buf *byte) (err error) {
    return nil
}

