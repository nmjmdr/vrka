package raft

type StartFollower struct {	
}

type ElectionAnounced struct {
}

type StartCandiate struct {
}

type Quit struct {
}

type VoteFrom struct {
	from string
	voteGranted bool
	termToUpdate uint64
}
