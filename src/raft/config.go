package raft

type Config interface {
	Peers() []Peer
}

type config struct {
	peers []Peer
}

func (c *config) Peers() []Peer {
	return c.peers
}
