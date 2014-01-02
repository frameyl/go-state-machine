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

const MSG_NAME_HELLO = "Hello"
const MSG_NAME_REQUEST = "Request"
const MSG_NAME_REPLY = "Reply"
const MSG_NAME_CONFIRM = "Confirm"
const MSG_NAME_CLOSE = "Close"

const (
    MSG_UNKNOWN   = iota
    MSG_HELLO
    MSG_REQUEST
    MSG_REPLY
    MSG_CONFIRM
    MSG_CLOSE
)

type MsgType int

func GetMsgNameByType(mtype MsgType) string {
    switch mtype {
    case MSG_UNKNOWN:
        return "Unknown"
    case MSG_HELLO:
        return MSG_NAME_HELLO
    case MSG_REQUEST:
        return MSG_NAME_REQUEST
    case MSG_REPLY:
        return MSG_NAME_REPLY
    case MSG_CONFIRM:
        return MSG_NAME_CONFIRM
    case MSG_CLOSE:
        return MSG_NAME_CLOSE
    default:
        return "ERROR"
    }
}

func ReadFieldString(pkt *bytes.Reader, offset int, length int) (string, error) {
    field := make([]byte, length)
    n, _ := pkt.ReadAt(field, int64(offset))
    if n != length {
        return "", fmt.Errorf("Packet is too short during reading field len %n, %v,", length, n)
    }

    index := bytes.IndexByte(field, 0)
    return string(field[:index]), nil
}

func IsSsmpPacket(pkt *bytes.Reader) (bool, error) {
    ProtoName, err := ReadFieldString(pkt, 0, LEN_PROTO_NAME)
    if err != nil {
        return false, err
    }

    if ProtoName == PROTO_NAME {
        return true, nil
    }
    return false, nil
}

func ReadMsgType(pkt *bytes.Reader) (MsgType, error) {
    msgTypeStr, err := ReadFieldString(pkt, OFF_MSG_NAME, LEN_MSG_NAME)
    if err != nil {
        return MSG_UNKNOWN, err
    }

    switch msgTypeStr {
    case MSG_NAME_HELLO:
        return MSG_HELLO, nil
    case MSG_NAME_REQUEST:
        return MSG_REQUEST, nil
    case MSG_NAME_REPLY:
        return MSG_REPLY, nil
    case MSG_NAME_CONFIRM:
        return MSG_CONFIRM, nil
    case MSG_NAME_CLOSE:
        return MSG_CLOSE, nil
    default:
        return MSG_UNKNOWN, nil
    }
}

func ReadMagicNum(pkt *bytes.Reader) (magic uint64, err error) {
    magicBytes := make([]byte, LEN_MAGIC_NUMBER)
    n, err := pkt.ReadAt(magicBytes, OFF_MAGIC_NUMBER)
    if n != LEN_MAGIC_NUMBER {
        return 0, fmt.Errorf("Found error during read magic number, %s", err)
    }

    magic = binary.BigEndian.Uint64(magicBytes)
    return
}


