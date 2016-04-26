package raft

// note: how to ensure a follower only votes once in a
// given term???

type follower struct {
}


func (f *follower) onElectionNotice(r *raftNode) state {
	//change role to candidate
	r.mutex.Lock()
	// increment the term and transition to a candidate
	r.currentTerm = r.currentTerm + 1
	r.role = Candidate
	r.mutex.Unlock()
	c := NewCandidate(r)	
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
