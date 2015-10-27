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
	
	return sg
}

func (sg *SessionGroupClient) Start() {
	
	
	return
}

func (sg *SessionGroupClient) Stop() {
	return
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
