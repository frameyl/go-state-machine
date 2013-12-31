package common

import (
    "container/list"
    "bytes"
    "fmt"
)

type DispatchMng struct {
    dList       list.List
}

// Dispatch counters
type DispatchCnt struct {
    Rx          uint
    Handled     uint
    Bypass      uint
    Discard     uint
}

type Dispatch   interface {
    // Main loop of the dispatch
    Handle(nextStep chan bytes.Reader) error
    // Return name of the dispatch
    Name() (string)
    // Get the buffer Channel of the Dispatch
    GetBufChan() chan bytes.Reader
    // Close the dispatch
    Close() error
    // Reset the dispatch
    Reset() error
    // Get Counters
    GetCnt() DispatchCnt
}

func (dq *DispatchMng)Add(dispatch Dispatch) (err error) {
    dq.dList.PushBack(dispatch)
    return nil
}

func (dq *DispatchMng)Remove(name string) (err error) {
    for e := dq.dList.Front(); e != nil; e = e.Next() {
        dispatch, _ := e.Value.(Dispatch)
        if dispatch.Name() == name {
            dq.dList.Remove(e)
            return nil
        }
    }

    return fmt.Errorf("Not found %s in Dispatcher Queue", name)
}

func (dq *DispatchMng)Start() (err error) {
    for e := dq.dList.Front(); e != nil; e = e.Next() {
        dispatch, _ := e.Value.(Dispatch)
        var nextStep chan bytes.Reader = nil
        if e.Next() != nil {
            tempDisp, _ := e.Next().Value.(Dispatch)
            nextStep = tempDisp.GetBufChan()
        }
        go dispatch.Handle(nextStep)
    }
    return nil
}

func (dq *DispatchMng)Stop() (err error) {
    for e := dq.dList.Front(); e != nil; e = e.Next() {
        dispatch, _ := e.Value.(Dispatch)
        dispatch.Close()
    }
    return nil
}

func (dq *DispatchMng)Handle(buf bytes.Reader) (err error) {
    e := dq.dList.Front()
    dispatch, _ := e.Value.(Dispatch)
    dispatch.GetBufChan() <- buf

    return nil
}

