package raft

import (
	"sync"
	"timerwrap"
	"time"
	"fmt"
)


type RaftNode interface {
	Role() Role
	Stop()
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
	lock *sync.Mutex
	state *State
	leaderState *LeaderState
	peers []Peer
	log Log
	monitor Monitor
	quit chan bool

	peerStore PeerStore
	stateStore StateStore
	trans Transport
	logStore LogStore
}

func (n *node) Role() Role {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.role
}

func (n *node) Stop() {
	n.quit <- true
}

func NewRaftNode(nodeId string,
	peerStore PeerStore,
	stateStore StateStore,
	trans Transport,
	logStore LogStore) RaftNode {
	
	n := ctor(nodeId,peerStore,stateStore,trans,logStore)
	
	n.monitor = NewMonitor(func(d time.Duration) timerwrap.TimerWrap {
		return timerwrap.NewTimer(d)
	})

	n.setup()
	
	return n
}

func ctor(
	nodeId string,
	peerStore PeerStore,
	stateStore StateStore,
	trans Transport,
	logStore LogStore) *node {

	n := new(node)
	
	n.id = nodeId
	n.peerStore = peerStore
	n.stateStore = stateStore
	n.trans = trans
	n.logStore = logStore
	n.lock = new(sync.Mutex)
	n.quit = make(chan bool)

	return n
}

func NewRaftNodeCustomTimer(nodeId string,
	peerStore PeerStore,
	stateStore StateStore,
	trans Transport,
	logStore LogStore,
	createTimer CreateTimer) RaftNode {
	
	n := ctor(nodeId,peerStore,stateStore,trans,logStore)
	
	
	n.monitor = NewMonitor(createTimer)
	n.setup()	
	return n
}




func (n *node) setup() {
	// start with Follower Role
	n.role = Follower
	n.loadPeers()
	n.watchElectionNotice()
	fmt.Println("Started watching for election timeout")
	n.monitor.Start()
}

func (n *node) watchElectionNotice() {
	go func() {
		for {
			fmt.Println("Waiting for election timeout")
			select {
			case <-n.monitor.ElectionNotice():
				// become the candidate
				fmt.Println("got timeout")
				n.lock.Lock()
				// vote for self and start as candidate
				n.role = Candidate
				n.lock.Unlock()			
			case <-n.quit:
				fmt.Println("got quit")
				n.monitor.Stop()
				break
			}
		}
	}()
}

func (n *node) loadPeers() {
	n.peers = n.peerStore.Get()
}


