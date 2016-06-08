package raft

import (
	"time"
	"fmt"
)

type nodeResponse struct {
	vres voteResponse
	err error
	delay time.Duration
}

type mockTransport struct {	
	responseMap map[string]nodeResponse
}

func newMockTransport() *mockTransport {
	m := new(mockTransport)
	m.responseMap = make(map[string]nodeResponse)
	return m
}

func (m *mockTransport) setResponse(id string,vr voteResponse,delay time.Duration,err error) {
	m.responseMap[id] = nodeResponse { vres : vr, err : err, delay : delay }
}

func (m *mockTransport) RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error) {

	responseParams,_ := m.responseMap[peer.Id]
	fmt.Printf("Delay is set to : %d\n",responseParams.delay)
	time.Sleep(responseParams.delay)
	fmt.Print("Mock Transport, returning: ")
	fmt.Println(responseParams.vres)
	return responseParams.vres,responseParams.err
	
}
