package raft

// note: how to ensure a follower only votes once in a
// given term???

type follower struct {
	r *raftNode
}


func NewFollower(r *raftNode) state {
	f := new(follower)
	f.r = r
	return f
}

func (f *follower) onElectionNotice() state {
	//change role to candidate
	
	f.r.mutex.Lock()

	// increment the term and transition to a candidate
	f.r.currentTerm = f.r.currentTerm + 1
	f.r.role = Candidate

	f.r.mutex.Unlock()

	c := NewCandidate(f.r)	
	return c
}




func (f *follower) onQuit() state  {
	return f
}


func (f *follower) onHeartbeat(beat Beat) state {
	return f
}
