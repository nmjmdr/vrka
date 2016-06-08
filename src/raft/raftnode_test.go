package raft

import (
	"testing"
	"timerwrap"
	"time"
	"sync"
	"fmt"
	"strconv"
)

func waitForStateChange(t *testing.T,n *node,wg *sync.WaitGroup, times int) {

	for times > 0 {
	select {
	case role,ok := <- n.stateChange:
		if ok {			
			t.Log(fmt.Sprintf("State change, current role %d",role))
		}
	}
		wg.Done()
		times = times - 1
	}
}

func Test_FollowerToCandidate(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	
	g := func(d time.Duration,nodeId string) timerwrap.TimerWrap {
		return timer
	}

	p := make([]Peer,1)
	p[0] = Peer{Id:"1",Address:""}
	config := NewMockConfig(p)
	transport := newInMemoryTransport()
	stable := newInMemoryStable()
	
	n := newNode("1",config,transport,g,stable)

	wg := sync.WaitGroup{}
	go waitForStateChange(t,n,&wg,1)

	wg.Add(1)
	start(n)
	wg.Wait()
	
	mockTimer,_ := timer.(*timerwrap.MockTimer)

	go waitForStateChange(t,n,&wg,1)
	wg.Add(1)
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}

	go waitForStateChange(t,n,&wg,1)
	wg.Add(1)
	stop(n)
	wg.Wait()	
	
}



func Test_CandidateToLeaderOneNode(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	
	g := func(d time.Duration,nodeId string) timerwrap.TimerWrap {
		return timer
	}

	p := make([]Peer,1)
	p[0] = Peer{Id:"1",Address:""}
	config := NewMockConfig(p)
	transport := newInMemoryTransport()
	stable := newInMemoryStable()
	
	n := newNode("1",config,transport,g,stable)

	wg := sync.WaitGroup{}
	go waitForStateChange(t,n,&wg,1)

	wg.Add(1)
	start(n)
	wg.Wait()
	
	mockTimer,_ := timer.(*timerwrap.MockTimer)

	go waitForStateChange(t,n,&wg,2)
	wg.Add(2) // two state changes
	mockTimer.Tick()
	wg.Wait()

	if n.role != Leader {
		t.Fatal("Should have been a leader")
	}
	

	go waitForStateChange(t,n,&wg,1)
	wg.Add(1)
	stop(n)
	wg.Wait()	
	

}

func getNNodes(n int) ([](*node),map[string]timerwrap.TimerWrap) {

	peers := make([]Peer,n)
	nodes := make([](*node),n)


	// construct the peers first
	for i:=0;i<n;i++ {
		iStr := strconv.Itoa(i)
		peers[i] = Peer{Id:iStr,Address:("addr:"+iStr)}
	}

	config := NewMockConfig(peers)

	transport := newMockTransport()

	timers := make(map[string]timerwrap.TimerWrap)
	
	for i :=0;i<n;i++ {
		iStr := strconv.Itoa(i)
		timers[iStr] = timerwrap.NewMockTimer()
	
		g := func(d time.Duration,nodeId string) timerwrap.TimerWrap {
			return timers[nodeId]
		}				
		
		stable := newInMemoryStable()
	
		nodes[i] = newNode(iStr,config,transport,g,stable)
	}

	return nodes,timers
}




func Test_CandidateToLeaderNNodes(t *testing.T) {

	numNodes := 3
	nodes,timers := getNNodes(numNodes)


	// set transport responses for nodes
	for i := 1;i<numNodes;i++ {
		iStr := strconv.Itoa(i)
		mt,_ := nodes[i].transport.(*mockTransport)
		mt.setResponse(iStr,voteResponse { voteGranted : true, from : iStr, termToUpdate : 0 }, 0, nil)
	}

	

	wgs := make([]sync.WaitGroup,numNodes)

	for i:=0;i<numNodes;i++ {
		wgs[i] = sync.WaitGroup{}
		go waitForStateChange(t,nodes[i],&wgs[i],1)
	}

	// start all
	for i := 0; i<numNodes; i++ {
		wgs[i].Add(1)
		start(nodes[i])
		wgs[i].Wait()
	}

	// node 0 transitions to candidate
	
	// wait for transition of node 0
	go waitForStateChange(t,nodes[0],&wgs[0],1)

	wgs[0].Add(1)
	mockTimer0,_ := timers["0"].(*timerwrap.MockTimer)
	fmt.Println("Tick to make node0 candidate")
	mockTimer0.Tick()
	wgs[0].Wait()
	
	
	if(nodes[0].role != Candidate) {
		t.Fatal("Should have been candidate")
	}

	go waitForStateChange(t,nodes[0],&wgs[0],1)
	wgs[0].Add(1)
	wgs[0].Wait()

	if(nodes[0].role != Leader) {
		t.Fatal("Node 0 should have been a leader")
	}
	
}


func Test_CandidteHigherTermDiscovered(t *testing.T) {

	numNodes := 3
	nodes,timers := getNNodes(numNodes)

	termToTest := uint64(2)
	termToUpdate := uint64(0)

	// set transport responses for nodes
	for i := 1;i<numNodes;i++ {
		iStr := strconv.Itoa(i)
		mt,_ := nodes[i].transport.(*mockTransport)
		vote := true
		if i == 1 {
			vote = false
			termToUpdate = termToTest
		} else {
			vote = true
			termToUpdate = 0
		}
		mt.setResponse(iStr,voteResponse { voteGranted : vote, from : iStr, termToUpdate : termToUpdate }, 0, nil)

	}	

	wgs := make([]sync.WaitGroup,numNodes)

	for i:=0;i<numNodes;i++ {
		wgs[i] = sync.WaitGroup{}
		go waitForStateChange(t,nodes[i],&wgs[i],1)
	}

	// start all
	for i := 0; i<numNodes; i++ {
		wgs[i].Add(1)
		start(nodes[i])
		wgs[i].Wait()
	}

	// node 0 transitions to candidate
	
	// wait for transition of node 0
	go waitForStateChange(t,nodes[0],&wgs[0],1)

	wgs[0].Add(1)
	mockTimer0,_ := timers["0"].(*timerwrap.MockTimer)
	fmt.Println("Tick to make node0 candidate")
	mockTimer0.Tick()
	wgs[0].Wait()
	
	
	if(nodes[0].role != Candidate) {
		t.Fatal("Should have been candidate")
	}

	go waitForStateChange(t,nodes[0],&wgs[0],1)
	wgs[0].Add(1)
	wgs[0].Wait()

	if(nodes[0].role != Follower) {
		t.Fatal("Node 0 should have been a follower")
	}

	if(nodes[0].currentTerm != termToTest) {
		t.Fatal("Node 0's term shoud have been updated")
	}
	
}

func Test_DelayedVotes(t *testing.T) {

	numNodes := 3
	nodes,timers := getNNodes(numNodes)

	// set transport responses for nodes
	for i := 1;i<numNodes;i++ {
		iStr := strconv.Itoa(i)
		mt,_ := nodes[i].transport.(*mockTransport)
		mt.setResponse(iStr,voteResponse { voteGranted : true, from : iStr, termToUpdate : 0}, time.Duration(100)*time.Millisecond, nil)

	}	

	wgs := make([]sync.WaitGroup,numNodes)

	for i:=0;i<numNodes;i++ {
		wgs[i] = sync.WaitGroup{}
		go waitForStateChange(t,nodes[i],&wgs[i],1)
	}

	// start all
	for i := 0; i<numNodes; i++ {
		wgs[i].Add(1)
		start(nodes[i])
		wgs[i].Wait()
	}

	// node 0 transitions to candidate
	
	// wait for transition of node 0
	go waitForStateChange(t,nodes[0],&wgs[0],1)

	wgs[0].Add(1)
	mockTimer0,_ := timers["0"].(*timerwrap.MockTimer)
	fmt.Println("Tick to make node0 candidate")
	mockTimer0.Tick()
	wgs[0].Wait()
	
	
	if(nodes[0].role != Candidate) {
		t.Fatal("Should have been candidate")
	}

	go waitForStateChange(t,nodes[0],&wgs[0],1)
	wgs[0].Add(1)
	wgs[0].Wait()

	if(nodes[0].role != Follower) {
		t.Fatal("Node 0 should have been a follower")
	}

	
}






