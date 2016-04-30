package raft

import (
	"testing"
	"time"
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

func Test_CandidateTransition(t *testing.T) {

	monitor := new(mockMonitor)
	// make it a buffered channel
	monitor.c = make(chan time.Time)

	config := NewConfig()
	
	node := NewRaftNode("id",monitor,config,nil)
	// signal it
	monitor.c <- time.Time{}
	
	node.Stop()
	if node.CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}
}

func Test_CandiateToFollower(t *testing.T) {

	monitor := new(mockMonitor)
	// make it a buffered channel
	monitor.c = make(chan time.Time)
	
	config := NewConfig()	
	node := NewRaftNode("id",monitor,config,nil)
	// signal it
	
	monitor.c <- time.Time{}	

	time.Sleep(2 * time.Millisecond)
	
	// sleep so that we give a chance for the node to transition to
	// to a candidate
	// need a better way to perform this - than sleeping, will figure out later
	if node.CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}


	beat := Beat{}
	// increment the beat term
	beat.Term = beat.Term + 1
	node.Heartbeat(beat)

	time.Sleep(2 * time.Millisecond)
	
	// sleep so that we give a chance for the node to transition to
	// to a follower
	
	if node.CurrentRole() != Follower {
		t.Fatal("should have been a follower")
	}
}



func Test_RejectBeat(t *testing.T) {

	monitor := new(mockMonitor)
	// make it a buffered channel
	monitor.c = make(chan time.Time)
	
	config := NewConfig()	
	node := NewRaftNode("id",monitor,config,nil)
	// signal it
	
	monitor.c <- time.Time{}	

	time.Sleep(2 * time.Millisecond)
	
	// sleep so that we give a chance for the node to transition to
	// to a candidate
	// need a better way to perfom this - than sleeping, will figure out later
	if node.CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}

	beat := Beat{}
	// increment the beat term
	// do not increment the term
	node.Heartbeat(beat)

	time.Sleep(2 * time.Millisecond)
	
	// should not have changed
	
	if node.CurrentRole() != Candidate {
		t.Fatal("should have been a candidate, as term was lesser")
	}
}
