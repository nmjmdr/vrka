package raft

import (
	"sync"
)


type Role int

const (
	Leader Role = 1
	Candidate Role = 2
	Follower Role = 3
)

type Beat struct {
	From string
	Term uint32
	Address string
}

type RaftNode interface {
	Stop()
	Heartbeat(beat Beat)
	CurrentRole() Role
}

type raftNode struct {
	id string
	role Role
	leader Peer

	mutex *sync.Mutex
	monitor Monitor
	quitCh chan bool
	heartbeatCh chan Beat
	wg *sync.WaitGroup
}

func NewRaftNode(id string,monitor Monitor) RaftNode {
	n := new(raftNode)
	n.id = id
	// start in follower role
	n.role = Follower
	n.mutex = &sync.Mutex{}
	n.quitCh = make(chan bool)
	n.heartbeatCh = make(chan Beat)
	n.monitor = monitor
	n.wg = &sync.WaitGroup{}
	// start watching for election notice
	n.watchForElection()
	return n
}

func (n *raftNode) Stop() {
	n.quitCh <- true
	n.wg.Wait()
}

func (n *raftNode) Heartbeat(beat Beat) {
	n.heartbeatCh <- beat
}

func (n *raftNode) CurrentRole() Role {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.role
}


func (n *raftNode) watchForElection() {
	go func() {
		n.wg.Add(1)
		defer n.wg.Done()
		quit := false
		for !quit {
			select {
			case _,ok := <-n.monitor.ElectionNotice():	
				if ok {
					// change role
					n.mutex.Lock()
					n.role = Candidate
					// start asking for votes and other things later here
					n.mutex.Unlock()
				}
			case beat,ok := <-n.heartbeatCh:
				if ok {
					n.monitor.Reset()
					n.mutex.Lock()
					// check the term and then accept the new leader
					// add the term verification rules later
					n.role = Follower
					n.leader = Peer{Id: beat.From, Address : beat.Address }
					n.mutex.Unlock()
				}
			case <-n.quitCh:
				n.monitor.Stop()
				quit = true
				break
			}
		}
	}()
}


