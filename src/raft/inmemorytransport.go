package raft


type inMemoryTransport struct {
	m map[string](*node)
}

func newInMemoryTransport() *inMemoryTransport {
	i :=new(inMemoryTransport)
	i.m = make(map[string](*node))
	return i
}


func (i *inMemoryTransport) setNode(id string,n *node) {
	i.m[id] = n
}



func (i *inMemoryTransport) RequestForVote(vreq voteRequest,peer Peer) (voteResponse,error) {
	n,ok := i.m[peer.Id]
	if !ok {
		panic("Peer not found in transport map")
	}
	
	vres,err := RequestForVote(n,vreq,peer)
	
	return vres,err
	
}
