package common

import "testing"

func TestSsmpDispatch(t *testing.T) {
    var dispMng DispatchMng
    ssmpDisp := NewSsmpDispatch("Ssmp Connector")

    dispMng.Add(ssmpDisp)
    dispMng.Start()

}


