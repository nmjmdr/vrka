package raft

import (
)

type candidate struct {
	r *raftNode
}


func NewCandidate(r *raftNode) *candidate {
	c := new(candidate)
	c.r = r
	return c
}

func (c *candidate) askForVotes() {

	for _,_ = range c.r.config.Peers() {
		
	}
}

func (c *candidate) onElectionNotice(r *raftNode) state {
	return c
}


func (c *candidate) onElectionResult(r *raftNode,elected bool) state  {
	return c
}


func (c *candidate) onQuit(r *raftNode) state {
	return c
}


func (c *candidate) onHeartbeat(r *raftNode,beat Beat) state {

	
	//change role to candidate
	r.mutex.Lock()
	r.role = Follower
	r.monitor.Reset()
	r.mutex.Unlock()

	// stop the election notices, reset the election timer
	f := new(follower)
	return f
}
