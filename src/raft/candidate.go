package raft

import (
	"sync"
)


type VoteRequest struct {
	term uint64
	candidateId string
	lastLogIndex uint64
	lastLogTerm uint64
}

type VoteResponse struct {
	termTpUpdate uint64
	voteGranted bool
}


type candidate struct {
	votes int
	mutex *sync.Mutex
	peerResponse chan VoteResponse
	r *raftNode
}


func NewCandidate(r *raftNode) *candidate {
	c := new(candidate)
	c.r = r
	c.mutex = &sync.Mutex{}
	c.peerResponse = make(chan VoteResponse)
	return c
}

func (c *candidate) askForVotes() {

	
	for _,_ = range c.r.config.Peers() {
		go func() {
			//vr := VoteRequest { term = c.r.term,c.r.id,0,0  }
			//c.r.transport.RequestForVote(
		}()		
	}
}

func (c *candidate) onElectionNotice(r *raftNode) state {
	return c
}


func (c *candidate) onElectionResult(r *raftNode,elected bool) state  {
	return c
}


func (c *candidate) onQuit(r *raftNode) state {
	return c
}


func (c *candidate) onHeartbeat(r *raftNode,beat Beat) state {

	
	//change role to candidate
	r.mutex.Lock()
	r.role = Follower
	r.monitor.Reset()
	r.mutex.Unlock()

	// stop the election notices, reset the election timer
	f := new(follower)
	return f
}
