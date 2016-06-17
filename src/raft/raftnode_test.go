package raft

import (
	"testing"
	"timerwrap"
	"time"
	"sync"
	"fmt"
	"strconv"
)

func listenToStateChange(t *testing.T,n *node,wg *sync.WaitGroup, times int) {

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

func makeNewNode(id string,electionTimer timerwrap.TimerWrap,heartbeatTimer timerwrap.TimerWrap,peers []Peer,transport Transport) *node {

	g := func(d time.Duration,nodeId string,forHint string) timerwrap.TimerWrap {
		if forHint == ForHeartbeatTimer {		
			fmt.Println("Returning heartbeat timer")
			return heartbeatTimer
		} 
		fmt.Println("Returning election timer")		
		return electionTimer
	}	

	config := NewMockConfig(peers)
	stable := newInMemoryStable()

	return newNode(id,config,transport,g,stable)
}

func waitForStop(t *testing.T,n *node, wg *sync.WaitGroup) {
	
	go listenToStateChange(t,n,wg,1)
	wg.Add(1)
	stop(n)
	wg.Wait()
}

func waitForStart(n *node, wg *sync.WaitGroup) {
	wg.Add(1)
	start(n)
	wg.Wait()
}


func Test_FollowerToCandidate(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	heartbeatTimer := timerwrap.NewMockTimer()

	peers := make([]Peer,1)
	peers[0] =  Peer{Id:"0",Address:""}
	n := makeNewNode("0",timer,heartbeatTimer,peers,newInMemoryTransport())

	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)

	mockTimer,_ := timer.(*timerwrap.MockTimer)

	go listenToStateChange(t,n,&wg,1)
	wg.Add(1)
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}

	waitForStop(t,n,&wg)		
}



func Test_CandidateToLeaderOneNode(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	heartbeatTimer := timerwrap.NewMockTimer()

	peers := make([]Peer,1)
	peers[0] =  Peer{Id:"0",Address:""}
	n := makeNewNode("0",timer,heartbeatTimer,peers,newInMemoryTransport())
	
	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)
	
	mockTimer,_ := timer.(*timerwrap.MockTimer)

	go listenToStateChange(t,n,&wg,2)
	wg.Add(2) // two state changes
	mockTimer.Tick()
	wg.Wait()

	if n.role != Leader {
		t.Fatal("Should have been a leader")
	}	

	waitForStop(t,n,&wg)
}



func Test_CandidateToLeaderNNodes(t *testing.T) {


	numNodes := 3

	timers := make([]timerwrap.TimerWrap,numNodes)
	heartbeatTimers := make([]timerwrap.TimerWrap,numNodes)
	
	peers := make([]Peer,numNodes)
	transport := newInMemoryTransport()
	wgs  := make([](*sync.WaitGroup),numNodes)
		
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
		timers[i] = timerwrap.NewMockTimer()
		heartbeatTimers[i] = timerwrap.NewMockTimer()
		wgs[i] = &sync.WaitGroup{}
	}
	

	nodes := make([](*node),numNodes)

	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		nodes[i] = makeNewNode(str,timers[i],heartbeatTimers[i],peers,transport)
		transport.setNode(str,nodes[i])
	}
	
	
	for i:=0;i<numNodes;i++ {
		go listenToStateChange(t,nodes[i],wgs[i],1)
		waitForStart(nodes[i],wgs[i])
	}
	
	mockTimer,_ := timers[0].(*timerwrap.MockTimer)

	go listenToStateChange(t,nodes[0],wgs[0],2)
	wgs[0].Add(2) // two state changes
	mockTimer.Tick()
	wgs[0].Wait()

	if nodes[0].role != Leader {
		t.Fatal("Should have been a leader")
	}	

	waitForStop(t,nodes[0],wgs[0])
}


func Test_CandidateTimedout(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	heartbeatTimer := timerwrap.NewMockTimer()

	numNodes := 3
	peers := make([]Peer,numNodes)
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
	}
	
	n := makeNewNode("0",timer,heartbeatTimer,peers,newMockTransport())
	
	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)
	

	// now the transport has to delay
	mockTransport,ok := n.transport.(*mockTransport)

	if !ok {
		panic("unable to case to mockTransport")
	}
	
	for i:=1;i<numNodes;i++ {
		str := strconv.Itoa(i)
		wg := &sync.WaitGroup{}
		vr := voteResponse{ termToUpdate:0,from:str,voteGranted:true }
		wg.Add(1)
		mockTransport.setVoteResponse(str,vr,wg,nil)
	}

	mockTimer,_ := timer.(*timerwrap.MockTimer)

	// make it a candidate
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}


	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	} 

	// now release the transport
	for i:=1;i<numNodes;i++ {
		str := strconv.Itoa(i)		
		mockTransport.releaseWait(str)
	}
	
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	wg.Wait()

	if n.role != Leader {
		t.Fatal("Should have been a leader")
	} 
}


func Test_CandidateHigherTerm(t *testing.T) {

	numNodes := 3

	timers := make([]timerwrap.TimerWrap,numNodes)
	heartbeatTimers :=  make([]timerwrap.TimerWrap,numNodes)
	
	peers := make([]Peer,numNodes)
	transport := newMockTransport()
	wgs  := make([](*sync.WaitGroup),numNodes)
		
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
		timers[i] = timerwrap.NewMockTimer()
		heartbeatTimers[i] = timerwrap.NewMockTimer()
		wgs[i] = &sync.WaitGroup{}
	}
	

	nodes := make([](*node),numNodes)


	termToUpdate := uint64(2)
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		nodes[i] = makeNewNode(str,timers[i],heartbeatTimers[i],peers,transport)
		vr := voteResponse{ termToUpdate:0,voteGranted:true,from:str }
		if i == 2 {
			vr.voteGranted = false
			vr.termToUpdate = termToUpdate
		}
		transport.setVoteResponse(str,vr,nil,nil)
	}
	
	
	for i:=0;i<numNodes;i++ {
		go listenToStateChange(t,nodes[i],wgs[i],1)
		waitForStart(nodes[i],wgs[i])
	}
	
	mockTimer,_ := timers[0].(*timerwrap.MockTimer)

	go listenToStateChange(t,nodes[0],wgs[0],2)
	wgs[0].Add(2) 
	mockTimer.Tick()
	wgs[0].Wait()
		

	waitForStop(t,nodes[0],wgs[0])

	if nodes[0].role != Follower {
		t.Fatal("Should have been a follower")
	}

	if nodes[0].currentTerm != termToUpdate {
		t.Fatal("Should have updated the term")
	}
	
}




func Test_AppendEntryHigherTerm(t *testing.T) {

	numNodes := 3

	timers := make([]timerwrap.TimerWrap,numNodes)
	heartbeatTimers := make([]timerwrap.TimerWrap,numNodes)
	
	peers := make([]Peer,numNodes)
	transport := newInMemoryTransport()
	wgs  := make([](*sync.WaitGroup),numNodes)
		
	
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
		timers[i] = timerwrap.NewMockTimer()
		heartbeatTimers[i] = timerwrap.NewMockTimer()
		wgs[i] = &sync.WaitGroup{}
	}
	

	nodes := make([](*node),numNodes)

	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		nodes[i] = makeNewNode(str,timers[i],heartbeatTimers[i],peers,transport)
		transport.setNode(str,nodes[i])
	}
	
	
	for i:=0;i<numNodes;i++ {
		go listenToStateChange(t,nodes[i],wgs[i],1)
		waitForStart(nodes[i],wgs[i])
	}
	
	mockTimer,_ := timers[0].(*timerwrap.MockTimer)

	go listenToStateChange(t,nodes[0],wgs[0],2)
	wgs[0].Add(2) // two state changes
	mockTimer.Tick()
	wgs[0].Wait()

	if nodes[0].role != Leader {
		t.Fatal("Should have been a leader")
	}

	// now send an Append Entry with higher term
	termToUpdate := uint64(3)
	entry := entry{ term : termToUpdate, from: "1" }
	AppendEntry(nodes[0],entry)


	go listenToStateChange(t,nodes[0],wgs[0],1)
	wgs[0].Add(1) 
	wgs[0].Wait()

	if nodes[0].role != Follower {
		t.Fatal("Should have been a follower")
	}

	if nodes[0].currentTerm != termToUpdate {
		t.Fatal("Should have updated the current term")
	}

	waitForStop(t,nodes[0],wgs[0])
}



func Test_AppendEntryReplyHigherTerm(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	heartbeatTimer := timerwrap.NewMockTimer()

	numNodes := 3
	peers := make([]Peer,numNodes)
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
	}
	
	n := makeNewNode("0",timer,heartbeatTimer,peers,newMockTransport())
	
	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)
	

	// now the transport has to delay
	mockTransport,ok := n.transport.(*mockTransport)

	if !ok {
		panic("unable to case to mockTransport")
	}

	termToUpdate := uint64(2)
	for i:=1;i<numNodes;i++ {
		str := strconv.Itoa(i)
		vr := voteResponse{ termToUpdate:0,from:str,voteGranted:true }
		mockTransport.setVoteResponse(str,vr,nil,nil)
		mockTransport.setAppendEntryResponse(str,false,termToUpdate)
	}

	mockTimer,_ := timer.(*timerwrap.MockTimer)
	
	// make it a candidate
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}
		
	go listenToStateChange(t,n,&wg,2)
	wg.Add(2) 
	wg.Wait()

	if n.role != Follower {
		t.Fatal("Should have been a follower")
	}

	if n.currentTerm != termToUpdate {
		t.Fatal("Should have had an updated term")
	}

}



func Test_AppendEntryReplyWithLowerTerm(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	heartbeatTimer := timerwrap.NewMockTimer()
	
	numNodes := 3
	peers := make([]Peer,numNodes)
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
	}
	
	n := makeNewNode("0",timer,heartbeatTimer,peers,newMockTransport())
	
	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)
	

	// now the transport has to delay
	mockTransport,ok := n.transport.(*mockTransport)

	if !ok {
		panic("unable to case to mockTransport")
	}
	
	for i:=1;i<numNodes;i++ {
		str := strconv.Itoa(i)
		vr := voteResponse{ termToUpdate:0,from:str,voteGranted:true }
		mockTransport.setVoteResponse(str,vr,nil,nil)
		mockTransport.setAppendEntryResponse(str,false,0)
	}

	mockTimer,_ := timer.(*timerwrap.MockTimer)
	
	// make it a candidate
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}
		
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	wg.Wait()

	if n.role != Leader {
		t.Fatal("Should have been a follower")
	}

	waitForStop(t,n,&wg)

	if n.role != Leader {
		t.Fatal("Should have been a follower")
	}
}


func Test_HeartbeatTimer(t *testing.T) {


	var timer timerwrap.TimerWrap

	numNodes := 3
	peers := make([]Peer,numNodes)
	for i:=0;i<numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}
	}
	

	
	var heartbeatTimer timerwrap.TimerWrap
	
	g := func(d time.Duration,nodeId string,forHint string) timerwrap.TimerWrap {
	
		if forHint == ForHeartbeatTimer {		
			fmt.Print("Returning heartbeat timer: ")
			heartbeatTimer = timerwrap.NewMockTimer()
			fmt.Println(heartbeatTimer)
			return heartbeatTimer
		} 

		timer = timerwrap.NewMockTimer()
		return timer
	}	

	config := NewMockConfig(peers)
	stable := newInMemoryStable()

	n := newNode("0",config,newMockTransport(),g,stable)
	
	wg := sync.WaitGroup{}
	
	go listenToStateChange(t,n,&wg,1)
	waitForStart(n,&wg)
	

	// now the transport has to delay
	mockTransport,ok := n.transport.(*mockTransport)

	if !ok {
		panic("unable to case to mockTransport")
	}

	callbackWg := &sync.WaitGroup{}
	callbackWg.Add((numNodes - 1)*2)
	cb := func(method string,params interface{}) {

		if method == "AppendEntry" {
			fmt.Println("Append Entry invoked")
			callbackWg.Done()
		} else {
			return
		}
	}
	
	
	for i:=1;i<numNodes;i++ {
		str := strconv.Itoa(i)	
		vr := voteResponse{ termToUpdate:0,from:str,voteGranted:true }	
		mockTransport.setVoteResponse(str,vr,nil,nil)
		mockTransport.setAppendEntryResponse(str,true,0)
		mockTransport.setCallback(str,cb)
	}

	mockTimer,_ := timer.(*timerwrap.MockTimer)

	// make it a candidate
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}

	
	go listenToStateChange(t,n,&wg,1)
	wg.Add(1) 
	wg.Wait()

	if n.role != Leader {
		t.Fatal("Should have been a leader")
	}

	mockHeartbeatTimer,_ := heartbeatTimer.(*timerwrap.MockTimer)
	fmt.Println(mockHeartbeatTimer)
	fmt.Println(mockTimer)
		
	mockHeartbeatTimer.Tick()

	callbackWg.Wait()	
	waitForStop(t,n,&wg)
}
