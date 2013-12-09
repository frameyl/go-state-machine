package ssmp

import "testing"

func TestEncodeFieldFixedLen(t *testing.T) {
    buf := []byte{}
    field := []byte{'s', 'a', 'm'}
    expected := []byte{'s', 'a', 'm', 0}
    var err error
    buf, err = EncodeFieldFixedLen(buf, field, 4) 
    if err != nil {
        t.Errorf("Encode field 1 failed, return a non-nil")
    } else if len(buf) != len(expected) {
        t.Errorf("Encode field 1 failed, expected len %v, actual len %v", len(expected), len(buf))
    } else {
        for i, element := range buf {
            if element != expected[i] {
                t.Errorf("Encode field 1 failed, expected[%v] %v, buf[%v] %v", i, expected[i], i, element)
            }
        }
    }

    buf = []byte{}
    field = []byte{'s', 'a', 'm', 'p', 'l', 'e'}
    expected = []byte{'s', 'a', 'm', 'p'}
    buf, err = EncodeFieldFixedLen(buf, field, 4) 
    if err != nil {
        t.Errorf("Encode field 2 failed, return a non-nil")
    } else if len(buf) != len(expected) {
        t.Errorf("Encode field 2 failed, expected len %v, actual len %v", len(expected), len(buf))
    } else {
        for i, element := range buf {
            if element != expected[i] {
                t.Errorf("Encode field 2 failed, expected[%v] %v, buf[%v] %v", i, expected[i], i, element)
            }
        }
    }
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
    if err != nil {
        t.Errorf("Encode header failed, return a non-nil")
    } else if len(buf) != len(expected) {
        t.Errorf("Encode header failed, expected len %v, actual len %v", len(expected), len(buf))
    } else {
        for i, element := range buf {
            if element != expected[i] {
                t.Errorf("Encode header failed, expected[%v] %v, buf[%v] %v", i, expected[i], i, element)
            }
        }
    }
}



