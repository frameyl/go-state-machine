package common

import "fmt"
import "encoding/binary"
import "bytes"

const LEN_PROTO_NAME = 16
const LEN_MSG_NAME = 8
const LEN_MAGIC_NUMBER = 8
const LEN_SERVER_ID = 32
const LEN_SSMP_HDR = LEN_PROTO_NAME + LEN_MSG_NAME + LEN_MAGIC_NUMBER + LEN_SERVER_ID
const LEN_SESSION_ID = 4

const OFF_MSG_NAME = LEN_PROTO_NAME
const OFF_MAGIC_NUMBER = OFF_MSG_NAME + LEN_MSG_NAME
const OFF_SERVER_ID = OFF_MAGIC_NUMBER + LEN_MAGIC_NUMBER
const OFF_SESSION_ID = OFF_SERVER_ID + LEN_SERVER_ID

const PROTO_NAME = "SSMPv1"

func ReadFieldString(pkt *bytes.Reader, offset int, length int) (string, error) {
    field := make([]byte, length)
    n, _ := pkt.ReadAt(field, int64(offset))
    if n != length {
        return "", fmt.Errorf("Packet is too short during reading field len %n, %v,", length, n)
    }

    index := bytes.IndexByte(field, 0)
    return string(field[:index]), nil
}

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
    index := bytes.IndexByte(buf[*offset:*offset+LEN_PROTO_NAME], 0)
    field := []byte{}
    hdr.Proto_name = append(field, buf[*offset:*offset+index]...)
    *offset += LEN_PROTO_NAME

    index = bytes.IndexByte(buf[*offset:*offset+LEN_PROTO_NAME], 0)
    field = []byte{}
    hdr.Msg_name = append(field, buf[*offset:*offset+index]...)
    *offset += LEN_MSG_NAME

    index = bytes.IndexByte(buf[*offset:*offset+LEN_PROTO_NAME], 0)
    field = []byte{}
    hdr.Magic_number = append(field, buf[*offset:*offset+LEN_MAGIC_NUMBER]...)
    *offset += LEN_MAGIC_NUMBER

    index = bytes.IndexByte(buf[*offset:*offset+LEN_PROTO_NAME], 0)
    field = []byte{}
    hdr.Server_id = append(field, buf[*offset:*offset+index]...)
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

