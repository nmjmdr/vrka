package raft

type Transport interface {
	RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error)
}


