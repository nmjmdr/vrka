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
 

type RaftNode struct {
	id string
	role Role
	mutex *sync.Mutex
	monitor Monitor
	quitCh chan bool
	wg *sync.WaitGroup
}

func NewRaftNode(id string,monitor Monitor) *RaftNode {
	n := new(RaftNode)
	n.id = id
	// start in follower role
	n.role = Follower
	n.mutex = &sync.Mutex{}
	n.quitCh = make(chan bool)
	n.monitor = monitor
	n.wg = &sync.WaitGroup{}
	// start watching for election notice
	n.watchForElection()
	return n
}

func (n *RaftNode) Stop() {
	n.quitCh <- true
	n.wg.Wait()
}


func (n *RaftNode) watchForElection() {
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
				
			case <-n.quitCh:
				n.monitor.Stop()
				quit = true
				break
			}
		}
	}()
}


