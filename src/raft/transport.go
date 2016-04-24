package raft

type Transport interface {
	RequestForVote(voteRequest VoteRequest) (term uint64,voteGranted bool)
}
