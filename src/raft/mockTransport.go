package raft

import (
	"fmt"
	"sync"
)

type nodeResponse struct {
	vres voteResponse
	err error
	wg *sync.WaitGroup
}

type mockTransport struct {	
	responseMap map[string]nodeResponse
}

func newMockTransport() *mockTransport {
	m := new(mockTransport)
	m.responseMap = make(map[string]nodeResponse)
	return m
}

func (m *mockTransport) setResponse(id string,vr voteResponse,wg *sync.WaitGroup,err error) {
	m.responseMap[id] = nodeResponse { vres : vr, err : err, wg : wg }
	fmt.Println("Map here: ")
	fmt.Println(m.responseMap)
}

func (m *mockTransport) RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error) {

	responseParams,ok := m.responseMap[peer.Id]

	if !ok {
		fmt.Println(m.responseMap)
		panic("response for peer not found in map")
	}
	
	if responseParams.wg != nil {
		responseParams.wg.Wait()
	}
	fmt.Print("Mock Transport, returning: ")
	fmt.Println(responseParams.vres)
	return responseParams.vres,responseParams.err	
}
