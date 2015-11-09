// ssmp_session_group.go
package common

import (
//"fmt"
)

type SessionGroupClient struct {
	sessions  []*SessionClient
	dispatch  *SsmpDispatch
	magicChan chan MagicReg
}

func NewSessionGroupClient(startid int, size int, dispatch *SsmpDispatch, outputChan chan []byte) *SessionGroupClient {
	sg := &SessionGroupClient{
		sessions:  make([]*SessionClient, size),
		dispatch:  dispatch,
		magicChan: make(chan MagicReg, size/50),
	}

	for i, _ := range sg.sessions {
		sg.sessions[i] = NewClientSession(startid+i, outputChan, sg.magicChan)
		go sg.sessions[i].RunClient()
	}

	go sg.Register()

	return sg
}

func (sg *SessionGroupClient) Start() {
	for _, s := range sg.sessions {
		s.CntlChan <- S_CMD_START
	}

	return
}

func (sg *SessionGroupClient) Stop() {
	for _, s := range sg.sessions {
		s.CntlChan <- S_CMD_STOP
	}

	return
}

func (sg *SessionGroupClient) Register() {
	for {
		reg := <-sg.magicChan
		if reg.BufChan == nil {
			sg.dispatch.Unregister(reg.Magic)
		} else {
			sg.dispatch.Register(reg.Magic, reg.BufChan)
		}
	}
}

type SessionGroupServer struct {
	sessions []*SessionServer
	dispatch *SsmpDispatch
}

func NewSessionGroupServer(startid int, size int, svrid string, dispatch *SsmpDispatch, outputChan chan []byte) *SessionGroupServer {
	sg := &SessionGroupServer{
		sessions: make([]*SessionServer, size),
		dispatch: dispatch,
	}

	for i, _ := range sg.sessions {
		sg.sessions[i] = NewServerSession(startid+i, uint32(startid+i), svrid, 0, outputChan)
	}

	return sg
}
