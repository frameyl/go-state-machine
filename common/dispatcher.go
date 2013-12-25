package common

import (
    "container/list"
    "bytes"
    "fmt"
)

type DispatcherQ struct {
    dList       list.List
}

type Dispatch   interface {
    Handle(chan bytes.Reader) (err error)
    Name() (string)
    GetBufChan() (chan bytes.Reader)
    GetCntlChan() (chan int)
}

func (dq *DispatcherQ)Add(dispatch Dispatch) (err error) {
    dq.dList.PushBack(dispatch)
    return nil
}

func (dq *DispatcherQ)Remove(name string) (err error) {
    for e := dq.dList.Front(); e != nil; e = e.Next() {
        dispatch, _ := e.Value.(Dispatch)
        if dispatch.Name() == name {
            dq.dList.Remove(e)
            return nil
        }
    }

    return fmt.Errorf("Not found %s in Dispatcher Queue", name)
}

func (dq *DispatcherQ)Start() (err error) {
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

func (dq *DispatcherQ)Stop() (err error) {
    for e := dq.dList.Front(); e != nil; e = e.Next() {
        dispatch, _ := e.Value.(Dispatch)
        dispatch.GetCntlChan() <- SSMP_DISP_CLOSE
    }
    return nil
}

func (dq *DispatcherQ)Handle(buf bytes.Reader) (err error) {
    e := dq.dList.Front()
    dispatch, _ := e.Value.(Dispatch)
    dispatch.GetBufChan() <- buf

    return nil
}

