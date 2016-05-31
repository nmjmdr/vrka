package raft

import (
	"fmt"
)

func start(n *node) {
	// raise initialize event
	fmt.Println("start invoked")
	n.eventCh <- new(Initialized)
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



/*
func startElection(n *node) {
	// request vote from all the peers in parallel
	peers := n.config.Peers()

	// set the lastLogIndex and lastLogTerm appropriately
	vreq := voteRequest{term:n.currentTerm, candidateId:n.id, lastLogIndex:0,lastLogTerm:0}
	
	for _,p := range peers {
		go func (p Peer) {
			vres,err := n.transport.RequestForVote(vreq,p)
			if err == nil {
				//response got
				voteFrom := VoteFrom { from:vres.from, termToUpdate:vres.termToUpdate, voteGranted:vres.voteGranted }

				n.eventCh <- voteFrom
			} else {
				// log the error?
			}
		}(p)
	}
}
*/
