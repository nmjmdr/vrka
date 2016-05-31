package raft

import (
	"testing"
	"timerwrap"
	"time"
	"sync"
	"fmt"
)

func waitForStateChange(t *testing.T,n *node,wg *sync.WaitGroup) {
	select {
	case role,ok := <- n.stateChange:
		if ok {			
			t.Log(fmt.Sprintf("State change, current role %d",role))
		}
	}
	wg.Done()
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
	n := newNode("1",config,transport,g)

	wg := sync.WaitGroup{}
	go waitForStateChange(t,n,&wg)

	wg.Add(1)
	start(n)
	wg.Wait()
	
	mockTimer,_ := timer.(*timerwrap.MockTimer)

	go waitForStateChange(t,n,&wg)
	wg.Add(1)
	mockTimer.Tick()
	wg.Wait()

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}

	go waitForStateChange(t,n,&wg)
	wg.Add(1)
	stop(n)
	wg.Wait()

	
	
	
}
