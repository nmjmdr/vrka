package raft

type RaftNode interface {
}

type Role uint32

const (
	Leader Role = 1
	Candidate Role = 2
	Follower Role = 3
)

type node struct {
	id string
	role Role
	state *State
	leaderState *LeaderState
	peers []Peer
	log Log

	peerStore PeerStore
	stateStore StateStore
	trans Transport
	logStore LogStore
}

func NewRaftNode(nodeId string,
	peerStore PeerStore,
	stateStore StateStore,
	trans Transport,
	logStore LogStore) RaftNode {
	
	n := new(node)
	n.id = nodeId
	n.peerStore = peerStore
	n.stateStore = stateStore
	n.trans = trans
	n.logStore = logStore

	// start with Follower Role
	n.role = Follower
	return n
}

func (n *node) loadPeers() {
	n.peers = n.peerStore.Get()
}


