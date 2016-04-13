package raft

import (
	"testing"
	"time"
	"timerwrap"
	"bufio"
	"bytes"
)

func Test_CandidateAfterTimeout(t *testing.T) {

	tw := timerwrap.NewMockTimer()
	peerStore := inMemoryPeerStore()
	f := func(d time.Duration) timerwrap.TimerWrap {
		return tw
	}
	node := NewRaftNode("node-1",peerStore,nil,nil,f)



	time.Sleep(1 * time.Second)

	// trigger election notice
	mockTimer,_ := tw.(*timerwrap.MockTimer)
	mockTimer.Tick()

	time.Sleep(1 * time.Second)
	
	if node.Role() != Candidate {
		t.Fail()
	}

	node.Stop()
	
}


func inMemoryPeerStore() PeerStore {
	p := make([]Peer,2)	
	p[0].Id = "1"
	p[1].Id = "2"

	buf := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buf)
	reader := bufio.NewReader(buf)

	peerstore := NewPeerStoreRW(reader,writer)
	peerstore.Set(p)
	writer.Flush()
	return peerstore
}
