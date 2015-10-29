// ssmp_session_group.go
package common

import (
	//"fmt"
)

type SessionGroupClient struct {
	sessions []*SessionClient
	dispatch *SsmpDispatch
}

func NewSessionGroupClient(startid int, size int, dispatch *SsmpDispatch) *SessionGroupClient {
	sg := &SessionGroupClient{
		sessions: make([]*SessionClient, size),
		dispatch: dispatch,
	}
	
	for i := range sg.sessions {
		sg.sessions[i] = NewClientSession(startid + i)
	}	
	
	go sg.Register()
	
	return sg
}

func (sg *SessionGroupClient) Start() {
	for _, s := range sg.sessions {
		go s.RunClient()
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
		reg := <- MagicChan
		if reg.BufChan == nil {
			sg.dispatch.Unregister(reg.Magic)
		} else {
			sg.dispatch.Register(reg.Magic, reg.BufChan)
		}
	}
}

type SessionGroupServer struct {
	sessions []SessionServer
	
}

func NewSessionGroupServer(size int) *SessionGroupServer {
	sg := &SessionGroupServer{
		sessions: make([]SessionServer, size),
	}
	
	return sg
}
