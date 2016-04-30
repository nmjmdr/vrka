package raft

// note: how to ensure a follower only votes once in a
// given term???

type leader struct {
	r *raftNode
}


func NewLeader(r *raftNode) state {
	l := new(leader)
	l.r = r
	return l
}

func (l *leader) onElectionNotice() state {
	return l
}




func (l *leader) onQuit() state  {
	return l
}


func (l *leader) onHeartbeat(beat Beat) state {
	return l
}
