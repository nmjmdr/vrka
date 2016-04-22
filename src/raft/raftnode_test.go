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
	
	node := NewRaftNode("id",monitor,nil,nil)
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
	
	node := NewRaftNode("id",monitor,nil,nil)
	// signal it]
	
	monitor.c <- time.Time{}	

	time.Sleep(2 * time.Millisecond)
	
	// sleep so that we give a chance for the node to transition to
	// to a candidate
	if node.CurrentRole() != Candidate {
		t.Fatal("should have been a candidate")
	}


	beat := Beat{}
	node.Heartbeat(beat)

	time.Sleep(2 * time.Millisecond)
	
	// sleep so that we give a chance for the node to transition to
	// to a follower
	
	if node.CurrentRole() != Follower {
		t.Fatal("should have been a follower")
	}
}
