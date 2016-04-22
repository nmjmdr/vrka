package raft

import (
)

type candidate struct {	
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
	r.mutex.Unlock()
		
	f := new(follower)
	return f
}
