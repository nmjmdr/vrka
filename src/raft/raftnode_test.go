package raft

import (
	"testing"
	"timerwrap"
	"time"
	"sync"
	"fmt"	
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

func makeNewNode(id string,timer timerwrap.TimerWrap) *node {

	g := func(d time.Duration,nodeId string) timerwrap.TimerWrap {
		return timer
	}	
	
	p := make([]Peer,1)
	p[0] = Peer{Id:id,Address:""}
	config := NewMockConfig(p)
	transport := newInMemoryTransport()
	stable := newInMemoryStable()

	return newNode(id,config,transport,g,stable)
}

func waitForStop(t *testing.T,n *node, wg *sync.WaitGroup) {
	
	go waitForStateChange(t,n,wg,1)
	wg.Add(1)
	stop(n)
	wg.Wait()
}


func Test_FollowerToCandidate(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	
	n := makeNewNode("0",timer)

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

	waitForStop(t,n,&wg)		
}



func Test_CandidateToLeaderOneNode(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	
	n := makeNewNode("0",timer)
	
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

	waitForStop(t,n,&wg)
}

