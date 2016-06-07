package raft

import (
	"fmt"
)


func RequestForVote(n *node,vreq voteRequest,peer Peer) (voteResponse,error) {

	// conditions to grant a vote:
	// request's term is greater than or equal to current term
	// not already voted in the current term
	// other conditons related to log - check it again

	 
	
	if vreq.term < n.currentTerm {
		return voteResponse { voteGranted:false,from:n.id,termToUpdate:n.currentTerm},nil
	}

	votedFor,ok := n.stable.Get(votedForKey)
	if !ok || votedFor == "" || votedFor == vreq.candidateId {
		// other checks: !!!
		// perform log checks

		// store that the vote was granted
		n.stable.Store(votedForKey,vreq.candidateId)
		// rasie votedFor event
		return voteResponse { voteGranted:true,from:n.id,termToUpdate:n.currentTerm},nil
		
	}

	
	return voteResponse { voteGranted:false,from:n.id,termToUpdate:n.currentTerm},nil
}

func start(n *node) {
	// raise initialize event
	fmt.Println("start invoked")
	n.eventCh <- new(StartFollower)
}

func stop(n *node) {
	fmt.Println("stop invoked")
	n.eventCh <- new(Quit)
}


func startElectionTimer(n *node) {
	n.electionTimer = n.getTimer(n.electionTimeout)

	fmt.Println("Will listen to election timer channel")
	
	go func() {
		select {
		case <-n.electionTimer.Channel():
			// election time out
			fmt.Println("got tick on election timer channel")
			n.eventCh <- new(ElectionAnounced)
		}
	}()
}





func startElection(n *node) {
	// request vote from all the peers in parallel
	peers := n.config.Peers()

	// set the lastLogIndex and lastLogTerm appropriately
	vreq := voteRequest{term:n.currentTerm, candidateId:n.id, lastLogIndex:0,lastLogTerm:0}

	go func() {
	// vote self
	selfVote := VoteFrom { from:n.id, termToUpdate:n.currentTerm, voteGranted:true}
	fmt.Println("self vote")
		n.eventCh <- &selfVote
	}()

	for _,p := range peers {
		if p.Id == n.id {
				continue
		}
		
		go func (p Peer) {
			fmt.Printf("Requesting vote from: %d",p.Id)
			vres,err := n.transport.RequestForVote(vreq,p)
			if err == nil {
				//response got
				voteFrom := VoteFrom { from:vres.from, termToUpdate:vres.termToUpdate, voteGranted:vres.voteGranted }

				n.eventCh <- &voteFrom
			} else {
				// log the error?
			}
		}(p)
	}
}

