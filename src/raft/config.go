package raft

type Config interface {
	Peers() []Peer
}

type mockConfig struct {
	peers []Peer
}

func NewMockConfig(peers []Peer) *mockConfig {
	c := new(mockConfig)
	c.peers = peers

	return c
}

func (c *mockConfig) Peers() []Peer {
	return c.peers
}
