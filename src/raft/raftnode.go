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


type state interface {
	onElectionNotice(r *raftNode) state
	onElectionResult(r *raftNode,elected bool) state
	onQuit(r *raftNode) state
	onHeartbeat(r *raftNode,beat Beat) state	
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
	votedFor string
	config *Config
	currentTerm uint32
	campaign Campaign

	mutex *sync.Mutex
	monitor Monitor
	quitCh chan bool
	heartbeatCh chan Beat
	wg *sync.WaitGroup

	state state
}





func NewRaftNode(id string,monitor Monitor,config *Config,campaign Campaign) RaftNode {
	r := new(raftNode)
	r.id = id
	// start in follower role
	r.role = Follower
	r.mutex = &sync.Mutex{}
	r.config = config
	r.campaign = campaign
	r.currentTerm = 0
	r.quitCh = make(chan bool)
	r.heartbeatCh = make(chan Beat)
	r.monitor = monitor
	r.wg = &sync.WaitGroup{}

	r.state = new(follower)
	
	// start watching for notices and heartbeats
	loop(r)
	return r
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






/* 
Follower state:
 Election Notice:    
      1. Stop election monitor
      2. Become candidate
      3. Start campaign
 Elected = YES Signal
      Invalid
 Elected = NO Signal
      Invalid
 Heartbeat
      1. Check the current term of the leader
      2. Acknowledge the other node as leader
      3. Reset the election monitor
 Quit
    1. Stop the election monitor
    2. Quit


Candidate state:
 Election Notice:
    Invalid   
 Elected = YES Signal
    1. Become the leader
    2. Start sending heartbeats (doubt: once elected as a leader - should we consider election timeout)
 Elected = NO Signal
    1. Nobody got elected, restart the campaign
 Heartbeat
    1. Check the current term of the leader
    2. Acknowledge the other node as leader
    3. Set to Follower
 Quit
    1. Stop the campaign
    2. Quit
 

Leader state:
 Election Notice:
    Invalid? (check doubt above)
 Elected = YES Signal
    Invalid
 Elected = NO Signal
    Invalid
 Heartbeat
    1. Check the current term of the leader
    2. Acknowledge the other node as leader
    3. Set to Follower
 Quit
    1. Stop the heartbeat signalling
    2. Quit
*/


func loop(node *raftNode) {	


	
	
	go func() {
		node.wg.Add(1)
		defer node.wg.Done()
 		quit := false
		for !quit {
			select {
			case _,ok := <-node.monitor.ElectionNotice():	
				if ok {
					node.state = node.state.onElectionNotice(node)
				}
			case beat,ok := <-node.heartbeatCh:
				if ok {
					node.state = node.state.onHeartbeat(node,beat)
				
				}
			case <-node.quitCh:			
				quit = true
				break
			}
		}
	}()
}


