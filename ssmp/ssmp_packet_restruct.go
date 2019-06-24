package common

import "fmt"
import "encoding/binary"
import "bytes"

type SsmpHeader struct {
	ProtoName		string `struct:"[16]byte"`
	MsgName 		string `struct:"[8]byte"`
	EpMagicNum		string `struct:[8]byte`
	ServerId		string `struct:[32]byte`
}


type SsmpWSid struct {
	SsmpHeader
	SessionID		uint32 `struct:uint32`
}

type SsmpHello struct {
	SsmpHeader
}

type SsmpRequest struct {
	SsmpHeader
}

type SsmpReply struct {
	SsmpWSid
}

type SsmpConfirm struct {
	SsmpWSid
}

type SsmpDisconnect struct {
	SsmpWSid
}

