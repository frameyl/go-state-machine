package ssmp

//import "fmt"
import "encoding/binary"

const LEN_PROTO_NAME = 16
const LEN_MSG_NAME = 8
const LEN_MAGIC_NUMBER = 8
const LEN_SERVER_ID = 32

type Packet_header struct {
    Proto_name      []byte
    Msg_name        []byte
    Magic_number    []byte
    Server_id       []byte
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

func EncodeFieldFixedLen(buf []byte, field []byte, length int) (result []byte, err error) {
    err = nil
    if len(field) >= length {
        result = append(buf, field[:length]...)
    } else {
        result = append(buf, field...)
        padding := make([]byte, length-len(field))
        result = append(result, padding...)
    }
    
    return
}

func EncodeFieldUint32(buf []byte, field uint32) (result []byte, err error) {
    sid := make([]byte, 4)
    err = nil
    binary.BigEndian.PutUint32(sid, field)
    result = append(buf, sid...)
    return
}

func (hdr Packet_header)Encode(buf []byte) (result []byte, err error) {
    result, err = EncodeFieldFixedLen(buf, hdr.Proto_name, LEN_PROTO_NAME)
    if err != nil {
        return
    }
    result, err = EncodeFieldFixedLen(result, hdr.Msg_name, LEN_MSG_NAME)
    if err != nil {
        return
    }
    result, err = EncodeFieldFixedLen(result, hdr.Magic_number, LEN_MAGIC_NUMBER)
    if err != nil {
        return
    }
    result, err = EncodeFieldFixedLen(result, hdr.Server_id, LEN_SERVER_ID)
    if err != nil {
        return
    }
    return
}

func (hdr Packet_hello)Encode(buf []byte) (result []byte, err error) {
    return hdr.Packet_header.Encode(buf)
}

func (hdr Packet_request)Encode(buf []byte) (result []byte, err error) {
    return hdr.Packet_header.Encode(buf)
}

func (hdr Packet_reply)Encode(buf []byte) (result []byte, err error) {
    result, err = hdr.Packet_header.Encode(buf)
    if err != nil {
        return buf, err
    }

    return EncodeFieldUint32(result, hdr.Session_id)
}

func (hdr Packet_confirm)Encode(buf []byte) (result []byte, err error) {
    result, err = hdr.Packet_header.Encode(buf)
    if err != nil {
        return buf, err
    }
    
    return EncodeFieldUint32(result, hdr.Session_id)
}



