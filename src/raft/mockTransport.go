package raft

import (
	"fmt"
	"sync"
)

type callback func(method string, params interface{})

type nodeResponse struct {
	vres voteResponse
	err error
	wg *sync.WaitGroup
	reply bool
	term uint64
	cb callback
	
}

type mockTransport struct {	
	responseMap map[string](*nodeResponse)
}

func newMockTransport() *mockTransport {
	m := new(mockTransport)
	m.responseMap = make(map[string](*nodeResponse))
	return m
}

func (m *mockTransport) setCallback(id string,cb callback) {
	nr,ok :=  m.responseMap[id]

	if !ok {
		m.responseMap[id] = &nodeResponse { cb : cb }
	} else {		
		nr.cb = cb		
	}
}

func (m *mockTransport) setVoteResponse(id string,vr voteResponse,wg *sync.WaitGroup,err error) {
	nr,ok :=  m.responseMap[id]

	if !ok {
		m.responseMap[id] = &nodeResponse { vres : vr, err : err, wg : wg }
	} else {		
		nr.vres = vr
		nr.err = err
		nr.wg = wg		
	}
	
}

func (m *mockTransport) releaseWait(id string) {
	responseParams,ok := m.responseMap[id]

	if !ok {
		fmt.Println(m.responseMap)
		panic("response for peer not found in map")
	}
	responseParams.wg.Done()
}

func (m *mockTransport) setAppendEntryResponse(id string,reply bool,term uint64) {
	nr,ok :=  m.responseMap[id]

	if !ok {
		m.responseMap[id] = &nodeResponse { reply:reply,term:term }
	} else {		
		nr.reply = reply
		nr.term = term
	}
	fmt.Print("set append entry response: ")
	fmt.Println(m.responseMap[id])
}


func (m *mockTransport) AppendEntry(entry entry,peer Peer) (bool,uint64) {

	responseParams,ok := m.responseMap[peer.Id]

	if !ok {
		fmt.Println(m.responseMap)
		panic("response for peer not found in map")
	}

	fmt.Printf("Peer : %s replying with: %t %d\n",peer.Id,responseParams.reply,responseParams.term)

	if responseParams.cb != nil {
		responseParams.cb("AppendEntry",entry)
	}
	
	return responseParams.reply,responseParams.term
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

	if responseParams.cb != nil {
		responseParams.cb("RequestForVote",vreq)
	}
	
	return responseParams.vres,responseParams.err	
}
