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

func Test_FollowerToCandidate(t *testing.T) {


	timer := timerwrap.NewMockTimer()
	
	g := func(d time.Duration) timerwrap.TimerWrap {
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
	
	g := func(d time.Duration) timerwrap.TimerWrap {
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
