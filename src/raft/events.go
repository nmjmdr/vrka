package raft

type StartFollower struct {	
}

type ElectionAnounced struct {
}

type StartCandiate struct {
}

type Quit struct {
}

type HigherTermDiscovered struct {
	term uint64
}

type VoteFrom struct {
	from string
	voteGranted bool
	termToUpdate uint64
}

type Append struct {
	from string
	term uint64
}

type AppendReply struct {
	reply bool
	from string
	term uint64
}

type TimeForHeartbeat struct {
}
