package common

import (
    "bytes"
//    "fmt"
)

type Fsm struct {
    Name    string  //unique name of the FSM
    State   string  //state of the FSM
    Magic   uint64  //Magic Number of FSM
    //packet recv channel, with packet goroutine
    BufChan  chan bytes.Reader
    //Control channel (start, reset, clean), with scheduler goroutine
    CntlChan chan int
}

const (
    FSM_CMD_START   = iota
    FSM_CMD_PAUSE
    FSM_CMD_RESET
    FSM_CMD_CLEAN
)

type stateFn func(*Fsm) stateFn

func (fsm *Fsm) run() {
    for state := Initial; state != nil; {
        state = state(fsm)
    }

    close(fsm.BufChan)
    close(fsm.CntlChan)
}

func Initial(fsm *Fsm) stateFn {
    // Send Hello packet periodicly, to Requesting phase once it get a hello from server with server id
    return Requesting
}

func Requesting(fsm *Fsm) stateFn {
    // Send Request packet periodicly, to Established phase once it get a reply from server
    return Established
}

func Established(fsm *Fsm) stateFn {
    // Doing nothing until it get a disconnect command
    return Disconnecting
}

func Disconnecting(fsm *Fsm) stateFn {
    // Send disconnect packet periodicly, to Initial state once it get a disconnect from server
    return Initial
}


