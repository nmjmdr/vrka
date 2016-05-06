package raft

import (
	"testing"
	"time"
	"fmt"
	"sync"
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

	
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i:=0;i<2;i++ {
		select {
		case role,_ := <- node.RoleChange():
			fmt.Printf("got role change: %d\n",role)
		}
		}
		wg.Done()
	}()


	// signal it
	monitor.c <- time.Time{}

	fmt.Println("waiting...")
	wg.Wait()	

	fmt.Println("stopping...")
	node.Stop()

	
	if node.CurrentRole() != leader {
		t.Fatal("should have been a leader")
	}

	
}


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
		vres := voteResponse{}
		vres.voteGranted = true
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}


	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i:=0;i<2;i++ {
		select {
		case role,_ := <- nodes[0].RoleChange():
			fmt.Printf("got role change: %d\n",role)
		}
		}
		wg.Done()
	}()

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	wg.Wait()
	

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
	

	if nodes[0].CurrentRole() != leader {
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
		vres := voteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	
	doneWg := sync.WaitGroup{}
	doneWg.Add(1)
	
	go func() {	
		for i:=0;i<2;i++ {
		select {
		case role,_ := <- nodes[0].RoleChange():
			fmt.Printf("got role change: %d\n",role)
		}
		}
		doneWg.Done()
	}()
	

	
	
	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	doneWg.Wait()

	if nodes[0].CurrentRole() != follower {
		t.Fatal("should have been a follower")
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
		vres := voteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransport(vres)

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for i:=0;i<2;i++ {
		select {
		case role,_ := <- nodes[0].RoleChange():
			if i ==0 && role != candidate {
				t.Fatal("should have been a candidate")
			} else if i == 1 && role != follower {
				t.Fatal("should have been follower")
			}
		}
		}
		wg.Done()
	}()

	// signal election notice for node id0
	monitors[0].c <- time.Time{}
	
	wg.Wait()

	
	wg.Add(1)
	go func() {	
		select {
		case role,_ := <- nodes[0].RoleChange():
			if role != candidate {
				t.Fatal("should have been a candidate")
			}		
		}
		wg.Done()
	}()

	// signal election notice for node id0
	monitors[0].c <- time.Time{}
	

	wg.Wait()

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}


func Test_TransitionWithDelayForVoting(t *testing.T) {

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
		vres := voteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransportWithDelay(vres,(500 * time.Millisecond))

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	
	doneWg := sync.WaitGroup{}
	doneWg.Add(1)
	
	go func() {		
		select {
		case role,_ := <- nodes[0].RoleChange():
			if role != candidate {
				t.Fatal("should have been a candidate")
			}
		
		}
		doneWg.Done()
	}()
	
	
	
	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	doneWg.Wait()


	doneWg.Add(1)
	
	go func() {		
		select {
		case role,_ := <- nodes[0].RoleChange():
			if role != follower {
				t.Fatal("should have been a follower")
			}
		
		}
		doneWg.Done()
	}()

	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	doneWg.Wait()

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}


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
		vres := voteResponse{}
		vres.voteGranted = false
		
		transports[i] = NewMockTransportWithDelay(vres,(500 * time.Millisecond))

		nodes[i] = NewRaftNode(fmt.Sprintf("id%d",i),monitors[i],config,transports[i])
	
	}

	
	doneWg := sync.WaitGroup{}
	doneWg.Add(1)
	
	go func() {		
		select {
		case role,_ := <- nodes[0].RoleChange():
			if role != candidate {
				t.Fatal("should have been a candidate")
			}
		
		}
		doneWg.Done()
	}()
	
	
	
	// signal election notice for node id0
	monitors[0].c <- time.Time{}

	doneWg.Wait()


	doneWg.Add(1)
	
	go func() {		
		select {
		case role,_ := <- nodes[0].RoleChange():
			if role != follower {
				t.Fatal("should have been a follower")
			}
		
		}
		doneWg.Done()
	}()

	// now signal a heartbeat from another node
	e := entry{}
	raft0,_ := nodes[0].(*raftNode)
	e.term = (raft0.term + 1)
	e.leaderId = "id1"
	nodes[0].Append(e)

	doneWg.Wait()

	for i:=0;i<n;i++ {
		nodes[i].Stop()
	}
}




