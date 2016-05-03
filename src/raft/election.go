package raft

import (
	"sync"
	"fmt"
)



type voteRequest struct {
	term uint64
	candidateId string
	lastLogIndex uint64
	lastLogTerm uint64
}

type voteResponse struct {
	termToUpdate uint64
	voteGranted bool
}


type Election interface {
	Start(r *raftNode)	
}

func NewElection() *election {

	e := new(election)
	return e
}

type election struct {	
}

func (e *election) Start(r *raftNode) {


	peers := r.config.Peers()


	fmt.Println("asking for votes")
	
	votes := e.askForVotes(peers,r)

	fmt.Printf("got votes: %d\n",votes)

	if votes >= (len(peers)/2 + 1) {
			// got elected
			r.electionResultCh <- true
	} else {
			// failed to get elected
			r.electionResultCh <- false
	}
}


func (c *election) askForVotes(peers []Peer,r *raftNode) int {
	
	vreq := voteRequest{}
	vreq.term = r.term
	vreq.candidateId = r.id
	vreq.lastLogIndex = 0 // set this later
	vreq.lastLogTerm = 0 // set this later


	//vote for self
	votes := 1
	
	wg := sync.WaitGroup{}
	// expect for the current node
	if len(peers) > 1 {
		wg.Add(len(peers)-1)
	}
		
	for i:=0;i<len(peers);i++ {
		if peers[i].Id == r.id {
			continue
		}

		go func(p Peer) {
			// try and optimize this by making the network calls async??
		
			vres,err := r.transport.RequestForVote(vreq)
			
			if err == nil && vres.voteGranted {
				votes++				
			}
			wg.Done()			
		}(peers[i])
	}


	wg.Wait()

	return votes
}
