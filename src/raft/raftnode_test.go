package raft

import (
	"testing"
	"timerwrap"
	"time"
)

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

	start(n)

	mockTimer,_ := timer.(*timerwrap.MockTimer)

	time.Sleep(200 * time.Millisecond)
	
	mockTimer.Tick()

	time.Sleep(200 * time.Millisecond)

	stop(n)
	
	wait(n)

	if n.role != Candidate {
		t.Fatal("Should have been a candidate")
	}
	
	
}
