package raft

type Transport interface {
	RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error)
	AppendEntry(entry entry,peer Peer) (bool,uint64)
}


