package common

type Fsm struct {
    Name    string  //unique name of the FSM
    State   string  //state of the FSM
    //packet recv channel, with packet goroutine
    //packet  chan Packet
    //mng channel (start, reset, clean), with scheduler goroutine
    //event   chan Event
    //timeout channel, with timeout goroutine
    //timeout chan Timeout
}


