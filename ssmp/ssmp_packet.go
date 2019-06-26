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

const MAGIC_NIL = 0
const USE_RESTRUCT = false

const (
	MSG_UNKNOWN = iota
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
		return fmt.Sprint("Unknow ", mtype)
	}
}

func GetMsgTypebyName(name string) MsgType {
	switch name {
	case MSG_NAME_HELLO:
		return MSG_HELLO
	case MSG_NAME_REQUEST:
		return MSG_REQUEST
	case MSG_NAME_REPLY:
		return MSG_REPLY
	case MSG_NAME_CONFIRM:
		return MSG_CONFIRM
	case MSG_NAME_CLOSE:
		return MSG_CLOSE
	default:
		return MSG_UNKNOWN
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
	if pkt.Len() < LEN_SSMP_HDR {
		return false, nil
	}

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

	return GetMsgTypebyName(msgTypeStr), nil
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

func ReadServerID(pkt *bytes.Reader) (string, error) {
	return ReadFieldString(pkt, OFF_SERVER_ID, LEN_SERVER_ID)
}

func ReadSessionID(pkt *bytes.Reader) (sid uint32, err error) {
	sidBytes := make([]byte, LEN_SESSION_ID)
	n, err := pkt.ReadAt(sidBytes, OFF_SESSION_ID)
	if n != LEN_SESSION_ID {
		return 0, fmt.Errorf("Found error during read session id, %s", err)
	}

	sid = binary.BigEndian.Uint32(sidBytes)
	return
}

func WriteFieldString(buf *bytes.Buffer, field string, length int) error {
	if length <= len(field) {
		_, err := buf.WriteString(field[:length])
		return err
	} else {
		buf.WriteString(field)
		_, err := buf.Write(make([]byte, length-len(field)))
		return err
	}

}

func WriteProtoName(buf *bytes.Buffer) error {
	return WriteFieldString(buf, PROTO_NAME, LEN_PROTO_NAME)
}

func WriteMsgName(buf *bytes.Buffer, mtype MsgType) error {
	return WriteFieldString(buf, GetMsgNameByType(mtype), LEN_MSG_NAME)
}

func WriteMagicNum(buf *bytes.Buffer, magic uint64) error {
	magicBytes := make([]byte, LEN_MAGIC_NUMBER)
	binary.BigEndian.PutUint64(magicBytes, magic)
	_, err := buf.Write(magicBytes)
	return err
}

func WriteServerID(buf *bytes.Buffer, svrid string) error {
	return WriteFieldString(buf, svrid, LEN_SERVER_ID)
}

func WritePacketHdr(buf *bytes.Buffer, mtype MsgType, magic uint64, svrid string) error {
	WriteProtoName(buf)
	WriteMsgName(buf, mtype)
	WriteMagicNum(buf, magic)
	WriteServerID(buf, svrid)
	return nil
}

func WriteSessionID(buf *bytes.Buffer, sid uint32) error {
	sidBytes := make([]byte, LEN_SESSION_ID)
	binary.BigEndian.PutUint32(sidBytes, sid)
	_, err := buf.Write(sidBytes)
	return err
}

func DumpSsmpPacket(pkt *bytes.Reader) string {
	var str string
	if isSsmp, _ := IsSsmpPacket(pkt); isSsmp {
		str = fmt.Sprintln("Protocol: ", PROTO_NAME)
		mType, _ := ReadMsgType(pkt)
		str += fmt.Sprintln("Type: ", GetMsgNameByType(mType))
		magic, _ := ReadMagicNum(pkt)
		str += fmt.Sprintf("Magic: 0x%X\n", magic)
		svrid, _ := ReadServerID(pkt)
		str += fmt.Sprintln("Server: ", svrid)

		if pkt.Len() >= LEN_SSMP_HDR+LEN_SESSION_ID {
			sid, _ := ReadSessionID(pkt)
			str += fmt.Sprintln("Session: ", sid)
		}
	} else {
		str = "Unknown Packet\n"
	}

	return str
}
