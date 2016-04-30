package raft

type Transport interface {
	RequestForVote(VoteRequest) (VoteResponse,error)
}
