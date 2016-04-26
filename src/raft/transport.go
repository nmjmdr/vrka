package raft

type Transport interface {
	RequestForVote(request VoteRequest,responseCh chan VoteResponse)
}
