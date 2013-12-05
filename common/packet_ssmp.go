package ssmp

//import "fmt"

const LEN_PROTO_NAME = 16
const LEN_MSG_NAME = 8
const LEN_MAGIC_NUMBER = 8
const LEN_SERVER_ID = 32

type Packet_header struct {
    Proto_name      [LEN_PROTO_NAME]byte
    Msg_name        [LEN_MSG_NAME]byte
    Magic_number    [LEN_MAGIC_NUMBER]byte
    Server_id       [LEN_SERVER_ID]byte
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

func (hdr Packet_header)encode(buf []byte) (err error) {
    buf = append(buf, hdr.Proto_name, hdr.Msg_name, hdr.Magic_number, hdr.Server_id)
    return nil
}

func (hdr Packet_hello)encode(buf []byte) (err error) {
    return hdr.Packet_header.encode(buf)
}

func (hdr Packet_request)encode(buf []byte) (err error) {
    return hdr.Packet_header.encode(buf)
}

func (hdr Packet_reply)encode(buf []byte) (err error) {
    err = hdr.Packet_header.encode(buf)
    if err != nil {
        return err
    }

    sid := [4]byte(hdr.Session_id)
    buf = append(buf, sid)
    return nil
}

func (hdr Packet_confirm)encode(buf []byte) (err error) {
    err = hdr.Packet_header.encode(buf)
    if err != nil {
        return err
    }

    sid := [4]byte(hdr.Session_id)
    buf = append(buf, sid)
    return nil
}



