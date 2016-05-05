package raft

type Transport interface {
	RequestForVote(voteRequest) (voteResponse,error)
}


