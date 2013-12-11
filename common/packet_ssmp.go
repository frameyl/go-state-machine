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
    Packet_reply
}

type Packet_disconnect struct {
    Packet_reply
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

func (reply Packet_reply)Encode(buf []byte) (result []byte, err error) {
    result, err = reply.Packet_header.Encode(buf)
    if err != nil {
        return buf, err
    }

    return EncodeFieldUint32(result, reply.Session_id)
}


func (hdr *Packet_header)Decode(buf []byte, offset *int) (err error) {
    field := []byte{}
    hdr.Proto_name = append(field, buf[*offset:*offset+LEN_PROTO_NAME]...)
    *offset += LEN_PROTO_NAME

    field = []byte{}
    hdr.Msg_name = append(field, buf[*offset:*offset+LEN_MSG_NAME]...)
    *offset += LEN_MSG_NAME

    field = []byte{}
    hdr.Magic_number = append(field, buf[*offset:*offset+LEN_MAGIC_NUMBER]...)
    *offset += LEN_MAGIC_NUMBER

    field = []byte{}
    hdr.Server_id = append(field, buf[*offset:*offset+LEN_SERVER_ID]...)
    *offset += LEN_SERVER_ID

    return nil
}

func (reply *Packet_reply)Decode(buf []byte, offset *int) (err error) {
    err = reply.Packet_header.Decode(buf, offset)
    if err != nil {
        return
    }

    reply.Session_id = binary.BigEndian.Uint32(buf[*offset:*offset+4])
    *offset += 4
    return nil
}

