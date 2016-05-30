package raft




type inMemoryTransport struct {
	n *node
}

func newInMemoryTransport() *inMemoryTransport {
	i :=new(inMemoryTransport)
	return i
}


func (i *inMemoryTransport) setNode(n *node) {
	i.n = n
}



func (i *inMemoryTransport) RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error) {
	//vres := i.n.RequestForVote(vreq)
	return voteResponse{},nil
	
}
