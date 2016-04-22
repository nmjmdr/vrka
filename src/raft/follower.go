package raft


type follower struct {
}


func (f *follower) onElectionNotice(r *raftNode) state {
	// change from follower to candidate
	// start the election process

	//change role to candidate
	r.mutex.Lock()
	r.role = Candidate
	r.mutex.Unlock()
	c := new(candidate)	
	return c
}


func (f *follower) onElectionResult(r *raftNode,elected bool) state {
	return f
}


func (f *follower) onQuit(r *raftNode) state  {
	return f
}


func (f *follower) onHeartbeat(r *raftNode,beat Beat) state {
	return f
}
