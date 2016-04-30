package raft

type Config interface {
	Peers() []Peer
}

type config struct {
	peers []Peer
}

func NewConfig() *config {
	c := new(config)
	c.peers = make([]Peer,1)

	return c
}

func (c *config) Peers() []Peer {
	return c.peers
}
