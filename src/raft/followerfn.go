package raft

// note: how to ensure a follower only votes once in a
// given term???

import (
	"fmt"
)

type followerScope struct {
}

func followerFn(r *raftNode,evt Evt)  {

	f := new(followerScope)
	
	fn := f.switchHandler(evt.Type())

	if fn != nil {
		fn(r,evt)
	} else {
		panic(fmt.Sprintf("got %d event while being a follower\n",evt.Type()))
	}
}

func (f *followerScope) switchHandler(t EvtType) stateFunction  {

	switch {		
	case t == ElectionTimeoutEvtType:
		return f.onElectionTimeout
	case t == HeartbeatEvtType:
		return f.onHeartbeat
	default:
		return nil
	}
}

func (f *followerScope) onHeartbeat(r *raftNode,evt Evt) {
	// got heartbeat
	// reset the election monitor
	r.monitor.Reset()
}


func (f *followerScope) onElectionTimeout(r *raftNode,evt Evt) {

	
	r.mutex.Lock()

	// increment the term and transition to a candidate
	r.term = r.term + 1
	fmt.Printf("Becoming a candidate: %s\n",r.id)
	r.role = candidate
	// change state function to candidateFn
	r.stateFn = candidateFn
	r.mutex.Unlock()

	// start the election here
	election := NewElection()
	go election.Start(r)
}
