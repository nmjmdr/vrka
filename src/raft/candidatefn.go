package raft

import (
	"fmt"
)

type candidateScope struct {
}

func candidateFn(r *raftNode,evt Evt)  {

	c := new(candidateScope)
	
	fn := c.switchHandler(evt.Type())

	if fn != nil {
		fn(r,evt)
	} else {
		panic(fmt.Sprintf("got %d event while being a candidate\n",evt.Type()))
	}

}

func (c *candidateScope) switchHandler(t EvtType) stateFunction  {
	switch {
	case t == ElectionResultEvtType:
		return c.onElectionResult
	case t == HeartbeatEvtType:
		return c.onHeartbeat
	default:
		return nil
	}
}


func (c *candidateScope) onHeartbeat(r *raftNode,evt Evt) {
	// reset the election monitor and move to a follower role
	r.mutex.Lock()
	r.role = follower
	r.mutex.Unlock()
	r.monitor.Reset()
}


func (c *candidateScope) onElectionResult(r *raftNode,evt Evt) {

	electionResult,_ := evt.(*electionResultEvt)

	r.mutex.Lock()
	// got elected
	if electionResult.elected {
		fmt.Printf("Becoming a leader: %s\n",r.id)
		r.role = leader
		r.stateFn = leaderFn
	} else {
		fmt.Printf("Did not get elected, becoming a follower: %s\n",r.id)
		r.role = follower
		r.stateFn = followerFn
		// reset the election timeout
		r.monitor.Reset()
	}	
	r.mutex.Unlock()
	
}
