package common

import "testing"
import "bytes"

var ssmp1 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'H', 'e', 'l', 'l', 'o', 0, 0, 0,
        0x11, 0x22, 0x33, 0x44, 0xaa, 0xbb, 0xcc, 0xdd,
        'N', 'o', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var ssmpPkt1 *bytes.Reader = bytes.NewReader(ssmp1)

var ssmp2 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'R', 'e', 'p', 'l', 'y', 0, 0, 0,
        0, 0, 0x1a, 0, 0x2b, 0, 0x3c, 0,
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
var ssmpPkt2 *bytes.Reader = bytes.NewReader(ssmp2)

var ssmpE1 []byte=[]byte{'S', 'S', 'M', 'P', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'r', 'e', 'p', 'l', 'y', '!', 0, 0,
        'm', 'a', 'g', '3', 'M', 'A', 'G', '4',
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
var ssmpEPkt1 *bytes.Reader = bytes.NewReader(ssmpE1)

var sample1 []byte=[]byte{'S', 'S', 'M', 'P', 'V', '1'}
var samplePkt1 *bytes.Reader = bytes.NewReader(sample1)

var sample2 []byte=[]byte{'S', 'S', 'M', 'P', 'V', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var samplePkt2 *bytes.Reader = bytes.NewReader(sample2)


func TestReadFieldString(t *testing.T) {
    ProtoName, _ := ReadFieldString(ssmpPkt1, 0, LEN_PROTO_NAME)
    if ProtoName != "SSMPv1" {
        t.Errorf("Read Field From ssmpPkt1 failed, exp %s, actual %s", ProtoName, "SSMPv1")
    }

    MsgName, _ := ReadFieldString(ssmpPkt1, OFF_MSG_NAME, LEN_MSG_NAME)
    if MsgName != "Hello" {
        t.Errorf("Read Field From ssmpPkt1 failed, exp %s, actual %s", MsgName, "hello")
    }

    _, err := ReadFieldString(samplePkt1, 2, 6)
    if err == nil {
        t.Errorf("Read Field should return a failure, but it didn't")
    }
}

func TestIsSsmpPacket(t *testing.T) {
    if IsSsmp, _ := IsSsmpPacket(ssmpPkt2); !IsSsmp {
        t.Errorf("SsmpPkt2 should be a SSMP packet, but it return false!")
    }

    if IsSsmp, _ := IsSsmpPacket(samplePkt1); IsSsmp {
        t.Errorf("SamplePkt1 should not be a SSMP packet!")
    }

    if IsSsmp, _ := IsSsmpPacket(samplePkt2); IsSsmp {
        t.Errorf("SamplePkt2 should not be a SSMP packet!")
    }
}

func TestReadMsgType(t *testing.T) {
    if mtype, _ := ReadMsgType(ssmpPkt1); mtype != MSG_HELLO {
        t.Errorf("SsmpPkt1 should be hello packet, return as a %s", GetMsgNameByType(mtype))
    }

    if mtype, _ := ReadMsgType(ssmpPkt2); mtype != MSG_REPLY {
        t.Errorf("SsmpPkt2 should be reply packet, return as a %s", GetMsgNameByType(mtype))
    }

    if mtype, _ := ReadMsgType(ssmpEPkt1); mtype != MSG_UNKNOWN {
        t.Errorf("SsmpEPkt1 should be unknown packet, return as a %s", GetMsgNameByType(mtype))
    }
}

func TestReadMagicNum(t *testing.T) {
    if magic, _ := ReadMagicNum(ssmpPkt1); magic != 0x11223344aabbccdd {
        t.Errorf("Read Magic Num failed, exp %X, actual %X", 0x11223344aabbccdd, magic)
    }

    if magic, _ := ReadMagicNum(ssmpPkt2); magic != 0x1a002b003c00 {
        t.Errorf("Read Magic Num failed, exp %X, actual %X", 0x1a002b003c00, magic)
    }
}

