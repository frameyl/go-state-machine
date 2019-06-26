package common

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/frameyl/log4go"
	"gopkg.in/restruct.v1"
)

type SsmpHeader struct {
	ProtoName  string `struct:"[16]byte"`
	MsgName    string `struct:"[8]byte"`
	EpMagicNum uint64 `struct:"uint64"`
	ServerId   string `struct:"[32]byte"`
}

type SsmpSid struct {
	SessionID uint32 `"struct:uint32"`
}

type SsmpHello struct {
	SsmpHeader
}

type SsmpRequest struct {
	SsmpHeader
}

type SsmpReply struct {
	SsmpHeader
	SsmpSid
}

type SsmpConfirm struct {
	SsmpHeader
	SsmpSid
}

type SsmpDisconnect struct {
	SsmpHeader
	SsmpSid
}

var LEN_SSMP_HDR_RESTRUCT, _ = restruct.SizeOf(SsmpHeader{})
var LEN_SSMP_SID_RESTRUCT, _ = restruct.SizeOf(SsmpSid{})

func WritePacketHdrRestruct(buf *bytes.Buffer, mtype MsgType, magic uint64, svrid string) (int, error) {
	header := SsmpHeader{PROTO_NAME, GetMsgNameByType(mtype), magic, svrid}
	bin, err := restruct.Pack(binary.BigEndian, &header)
	if err != nil {
		log4go.Error("Pack Header failed %s, %d, %s", mtype, magic, svrid)
		return 0, err
	}
	bytes, err := buf.Write(bin)
	if err != nil {
		log4go.Error("Write Header failed %s, %d, %s", mtype, magic, svrid)
	}
	return bytes, err
}

func ReadPacketHdrRestruct(pkt *bytes.Reader) (MsgType, uint64, string, error) {
	field := make([]byte, LEN_SSMP_HDR_RESTRUCT)
	n, err := pkt.Read(field)
	if n != LEN_SSMP_HDR_RESTRUCT {
		return MSG_UNKNOWN, 0, "", fmt.Errorf("Packet is too short during reading field len %d, %v,", LEN_SSMP_HDR_RESTRUCT, n)
	}
	var header SsmpHeader
	err = restruct.Unpack(field, binary.BigEndian, &header)
	if err != nil {
		log4go.Error("Read Header failed %v", field)
		return MSG_UNKNOWN, 0, "", err
	}
	return GetMsgTypebyName(header.MsgName), header.EpMagicNum, header.ServerId, nil
}

func WriteSessionIDRestruct(buf *bytes.Buffer, sid uint32) (int, error) {
	pkt := SsmpSid{SessionID: sid}
	bin, err := restruct.Pack(binary.BigEndian, &pkt)
	if err != nil {
		log4go.Error("SessionID failed %d", sid)
		return 0, err
	}
	bytes, err := buf.Write(bin)
	if err != nil {
		log4go.Error("SessionID failed %d", sid)
	}
	return bytes, err
}

func ReadSessionIDRestruct(pkt *bytes.Reader) (uint32, error) {
	field := make([]byte, LEN_SSMP_SID_RESTRUCT)
	n, err := pkt.Read(field)
	if n != LEN_SSMP_SID_RESTRUCT {
		return 0, fmt.Errorf("Packet is too short during reading field len %d, %v,", LEN_SSMP_SID_RESTRUCT, n)
	}
	var sessionid SsmpSid
	err = restruct.Unpack(field, binary.BigEndian, &sessionid)
	if err != nil {
		log4go.Error("Read Header failed %v", field)
		return 0, err
	}
	return sessionid.SessionID, nil
}
