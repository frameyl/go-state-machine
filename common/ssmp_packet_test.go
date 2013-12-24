package common

import "testing"

func TestEncodeFieldFixedLen(t *testing.T) {
    buf := []byte{}
    field := []byte{'s', 'a', 'm'}
    expected := []byte{'s', 'a', 'm', 0}
    var err error
    buf, err = EncodeFieldFixedLen(buf, field, 4)
    checkEncodedBuf(err, buf, expected, "Encode field 1", t)

    buf = []byte{}
    field = []byte{'s', 'a', 'm', 'p', 'l', 'e'}
    expected = []byte{'s', 'a', 'm', 'p'}
    buf, err = EncodeFieldFixedLen(buf, field, 4)
    checkEncodedBuf(err, buf, expected, "Encode field 2", t)
}

func TestPacketheaderEncode1(t *testing.T) {
    hdr := Packet_header{}
    hdr.Proto_name = []byte("ssmp v1")
    hdr.Msg_name = []byte("hello")
    hdr.Magic_number = []byte("mag1MAG2")
    hdr.Server_id = []byte("No Name")
    expected := []byte{'s', 's', 'm', 'p', ' ', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'h', 'e', 'l', 'l', 'o', 0, 0, 0,
        'm', 'a', 'g', '1', 'M', 'A', 'G', '2',
        'N', 'o', ' ', 'N', 'a', 'm', 'e', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
    buf := []byte{}
    var err error

    buf, err = hdr.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode header", t)

    pktHello := Packet_hello{hdr}
    buf = []byte{}
    buf, err = pktHello.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode hello", t)

    pktRequest := Packet_request{hdr}
    buf = []byte{}
    buf, err = pktRequest.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode request", t)
}

func TestEncodeFieldUint32(t *testing.T) {
    buf := []byte{}
    field := uint32(0x1f2f3f4f)
    expected := []byte{0x1f, 0x2f, 0x3f, 0x4f}
    var err error

    buf, err = EncodeFieldUint32(buf, field)
    checkEncodedBuf(err, buf, expected, "Encode uint32", t)
}

func TestPacketReplyEncode(t *testing.T) {
    pktReply := Packet_reply{}
    pktReply.Proto_name = []byte("ssmp v1")
    pktReply.Msg_name = []byte("reply")
    pktReply.Magic_number = []byte("mag3MAG4")
    pktReply.Server_id = []byte("Slient Lamp")
    pktReply.Session_id = uint32(0x1a2b3c4d)
    expected := []byte{'s', 's', 'm', 'p', ' ', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'r', 'e', 'p', 'l', 'y', 0, 0, 0,
        'm', 'a', 'g', '3', 'M', 'A', 'G', '4',
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
    buf := []byte{}
    var err error

    buf, err = pktReply.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode reply", t)

    pktConfirm := Packet_confirm{pktReply}
    buf = []byte{}
    buf, err = pktConfirm.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode confirm", t)

    pktDisconnect := Packet_disconnect{pktReply}
    buf = []byte{}
    buf, err = pktDisconnect.Encode(buf)
    checkEncodedBuf(err, buf, expected, "Encode disconnect", t)
}

func checkEncodedBuf(err error, buf, expected []byte, errStr string, t *testing.T) {
    if err != nil {
        t.Errorf("%s failed, return a non-nil", errStr)
    } else if len(buf) != len(expected) {
        t.Errorf("%s failed, expected len %v, actual len %v", errStr, len(expected), len(buf))
    } else {
        for i, element := range buf {
            if element != expected[i] {
                t.Errorf("%s failed, expected[%v] %v, buf[%v] %v", errStr, i, expected[i], i, element)
            }
        }
    }
}

func TestDecodePacketHeader(t *testing.T) {
    buf := []byte{'s', 's', 'm', 'p', ' ', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'h', 'e', 'a', 'd', 'e', 'r', 0, 0,
        'm', 'a', 'g', '7', 'M', 'A', 'G', '8',
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 }
    offset := int(0)
    hdr := Packet_header{}
    expected := Packet_header{}
    expected.Proto_name = []byte("ssmp v1")
    expected.Msg_name = []byte("header")
    expected.Magic_number = []byte("mag7MAG8")
    expected.Server_id = []byte("Slient Lamp")
    var err error

    err = hdr.Decode(buf, &offset)
    if err != nil {
        t.Errorf("Decode header failed, return a non-nil")
    }

    checkFieldDecode(t, hdr.Proto_name, expected.Proto_name, "Decode Packet_header.Proto_name failed")
    checkFieldDecode(t, hdr.Msg_name, expected.Msg_name, "Decode Packet_header.Msg_name failed")
    checkFieldDecode(t, hdr.Magic_number, expected.Magic_number, "Decode Packet_header.Magic_number failed")
    checkFieldDecode(t, hdr.Server_id, expected.Server_id, "Decode Packet_header.Server_id failed")
}

func TestDecodePacketReply(t *testing.T) {
    buf := []byte{'s', 's', 'm', 'p', ' ', 'v', '1', 0, 0, 0, 0, 0, 0, 0, 0, 0,
        'h', 'e', 'a', 'd', 'e', 'r', 0, 0,
        'm', 'a', 'g', '7', 'M', 'A', 'G', '8',
        'S', 'l', 'i', 'e', 'n', 't', ' ', 'L', 'a', 'm', 'p', 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0x1a, 0x2b, 0x3c, 0x4d }
    offset := int(0)
    reply := Packet_reply{}
    expected := Packet_reply{}
    expected.Proto_name = []byte("ssmp v1")
    expected.Msg_name = []byte("header")
    expected.Magic_number = []byte("mag7MAG8")
    expected.Server_id = []byte("Slient Lamp")
    expected.Session_id = 0x1a2b3c4d
    var err error

    err = reply.Decode(buf, &offset)
    if err != nil {
        t.Errorf("Decode reply failed, return a non-nil")
    }

    checkFieldDecode(t, reply.Proto_name, expected.Proto_name, "Decode Packet_reply.Proto_name failed")
    checkFieldDecode(t, reply.Msg_name, expected.Msg_name, "Decode Packet_reply.Msg_name failed")
    checkFieldDecode(t, reply.Magic_number, expected.Magic_number, "Decode Packet_reply.Magic_number failed")
    checkFieldDecode(t, reply.Server_id, expected.Server_id, "Decode Packet_reply.Server_id failed")

    if (expected.Session_id != reply.Session_id) {
        t.Errorf("Decode Packet_repy.Session_id failed, expected %v, actaul %v", expected.Session_id, reply.Session_id)
    }
}

func checkFieldDecode(t *testing.T, field, expected []byte, errStr string) {
    if len(field) != len(expected) {
        t.Errorf("%s failed, expected len %v, actual %v", errStr, len(expected), len(field))
    } else {
        for i, element := range field {
            if element != expected[i] {
                t.Errorf("%s failed, expected[%v] %v, actual[%v] %v", errStr, i, expected[i], i, element)
            }
        }
    }
}

