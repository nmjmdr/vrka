package raft

import (
	"testing"
	"time"
	//"fmt"
)

type mockMonitor struct {
	c chan time.Time
}

func (m *mockMonitor) Stop() bool {
	return true
}

func (m *mockMonitor) ElectionNotice() (<-chan time.Time) {
	return m.c
}

func (m *mockMonitor) Reset() {
}


func Test_SinglePeerTransition(t *testing.T) {

	monitor := new(mockMonitor)
	monitor.c = make(chan time.Time)

	peers := make([]Peer,1)
	peers[0].Id = "id1"
	config := NewMockConfig(peers)

	vres := voteResponse{}
	vres.voteGranted = true
	
	transport := NewMockTransport(vres)
	
	node := NewRaftNode("id1",monitor,config,transport)
	// signal it
	monitor.c <- time.Time{}

	
	// there should 2 transitions
	for i:=0;i<2;i++ {
		select {
		case <- node.RoleChange():
		}
	}


	node.Stop()

	if node.CurrentRole() != leader {
		t.Fatal("should have been a leader")
	}

	
}

/*
func Test_ThreePeersTransition(t *testing.T) {

	n := 3

	//setup peers
	peers := make([]Peer,n)
	for i:=0;i<n;i++ {
		peers[i].Id = fmt.Sprintf("id%d",i)
	}
	config := NewMockConfig(peers)
	
	monitors := make([](*mockMonitor),n)
	transports := make([](*mockTransport),n)
	nodes := make([]RaftNode,n)
		
	for i:=0;i<n;i++ {
		monitors[i] = new(mockMonitor)
		monitors[i].c = make(chan time.Time)

		// give the vote
		vres := VoteResponse{}
		vres.voteGranted = true
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	//wait for role change

	select {
	case <- nodes[0].RoleChange():
	}
	

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
	

	if nodes[0].CurrentRole() != Leader {
		t.Fatal("should have been a leader")
	}
}


func Test_ThreePeersTransitionNoVotes(t *testing.T) {

	n := 3

	//setup peers
	peers := make([]Peer,n)
	for i:=0;i<n;i++ {
		peers[i].Id = fmt.Sprintf("id%d",i)
	}
	config := NewMockConfig(peers)

		
	monitors := make([](*mockMonitor),n)
	transports := make([](*mockTransport),n)
	nodes := make([]RaftNode,n)
		
	for i:=0;i<n;i++ {
		monitors[i] = new(mockMonitor)
		monitors[i].c = make(chan time.Time)

		// give the vote
		vres := VoteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	select {
	case <- nodes[0].RoleChange():
	}	

	if nodes[0].CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}



func Test_ThreePeersTransitionRelection(t *testing.T) {

	n := 3

	//setup peers
	peers := make([]Peer,n)
	for i:=0;i<n;i++ {
		peers[i].Id = fmt.Sprintf("id%d",i)
	}
	config := NewMockConfig(peers)

		
	monitors := make([](*mockMonitor),n)
	transports := make([](*mockTransport),n)
	nodes := make([]RaftNode,n)
		
	for i:=0;i<n;i++ {
		monitors[i] = new(mockMonitor)
		monitors[i].c = make(chan time.Time)

		// give the vote
		vres := VoteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	select {
	case <- nodes[0].RoleChange():
	}

	//fmt.Printf("The role is: %d\n",nodes[0].CurrentRole())

	if nodes[0].CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}

	// signal election notice for node id0, so that it now transitions
	// to a follower
	monitors[0].c <- time.Time{}

	select {
	case <- nodes[0].RoleChange():
	}

	if nodes[0].CurrentRole() != Follower {
		t.Fatal("should have been a follower")
	}

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}

/*
func Test_ThreePeersTransitionOtherNodeHeartbeat(t *testing.T) {

	n := 3

	//setup peers
	peers := make([]Peer,n)
	for i:=0;i<n;i++ {
		peers[i].Id = fmt.Sprintf("id%d",i)
	}
	config := NewMockConfig(peers)

		
	monitors := make([](*mockMonitor),n)
	transports := make([](*mockTransport),n)
	nodes := make([]RaftNode,n)
		
	for i:=0;i<n;i++ {
		monitors[i] = new(mockMonitor)
		monitors[i].c = make(chan time.Time)

		// give the vote
		vres := VoteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	select {
	case <- nodes[0].RoleChange():
	}
	
	if nodes[0].CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}

	//  node 1 was elected as leader, and it sends heartbeat
	// make node 0 receive the heartbeat
	beat := Beat{}
	beat.From = "id1"
	raft0,_ := nodes[0].(*raftNode)
	beat.Term = raft0.currentTerm + 1
	nodes[0].Heartbeat(beat)

	select {
	case <- nodes[0].RoleChange():
	}
	
	
	// now node 0 should transition to a follower
	if nodes[0].CurrentRole() != Follower {
		t.Fatal("should have been a follower")
	}

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}
*/



