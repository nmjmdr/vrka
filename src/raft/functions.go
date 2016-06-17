package raft

import (
	"fmt"
	"timerwrap"
)

/*
There are three entry points where a node communicates with other nodes:

 - Answering - RequestForVote
 - Response from - request for voe that the node sent
 - Append entry

 - if term > currentTerm, set currentTerm = term
   and set as follower
   event : higherTermDiscovered

*/


func RequestForVote(n *node,vreq voteRequest,peer Peer) (voteResponse,error) {

	// conditions to grant a vote:
	// request's term is greater than or equal to current term
	// not already voted in the current term
	// other conditons related to log - check it later

	if vreq.term <= n.currentTerm {
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

	n.ifHigherTerm(vreq.term)
	
	return voteResponse { voteGranted:false,from:n.id,termToUpdate:n.currentTerm},nil
}

func (n *node) ifHigherTerm(term uint64) {
	if term > n.currentTerm {
		go func(term uint64) {
			higherTermEvent := new(HigherTermDiscovered)
			higherTermEvent.term = term
			n.eventCh <- higherTermEvent
		}(term)
	}
}

func start(n *node) {
	// raise initialize event
	fmt.Println("start invoked")
	n.eventCh <- new(StartFollower)
}

func stop(n *node){
	fmt.Println("stop invoked")
	n.eventCh <- new(Quit)
}


func startElectionTimer(n *node) {

	if n.electionTimer == nil {
		n.electionTimer = n.getTimer(n.electionTimeout,n.id,ForElectionTimer)
	} else {
		n.electionTimer.Reset(n.electionTimeout)
	}

	fmt.Println("Will listen to election timer channel")
	
	go func(timer timerwrap.TimerWrap) {
		select {
		case _,ok := <-timer.Channel():

			if ok {
			// election time out
			fmt.Printf("%s: got tick on election timer channel\n",n.id)
				n.eventCh <- new(ElectionAnounced)
			}
		}
	}(n.electionTimer)
}


func AppendEntry(n *node,entry entry) (reply bool,termToUpdate uint64)  {


	fmt.Printf("%s - got Append Entry from: %s\n",n.id,entry.from)
	
	if n.currentTerm > entry.term {
		return false,n.currentTerm
	}

	n.ifHigherTerm(entry.term)

	
	go func() {
		append := Append{ from : entry.from, term : entry.term }
		fmt.Printf("%s - sending event Append, received from: %s\n",n.id,append.from)
		n.eventCh <- &append
	}()	

	return true,entry.term
}

func stopHeartbeatTimer(n *node) {

	fmt.Printf("%s - Stopping heartbeat timer\n",n.id)
	if n.heartbeatTimer != nil {
		n.heartbeatTimer.Stop()
		fmt.Printf("%s - Stopped heartbeat timer\n",n.id)
	}
}

func stopElectionTimer(n *node) {

	fmt.Printf("%s - Stopping election timer\n",n.id)
	if n.electionTimer != nil {
		n.electionTimer.Stop()
		fmt.Printf("%s - Stopped election timer\n",n.id)
	}
}



func startHeartbeatTimer(n *node) {

	if n.heartbeatTimer == nil {
		n.heartbeatTimer = n.getTimer(n.heartbeatTimeout,n.id,ForHeartbeatTimer)
	} else {
		n.heartbeatTimer.Reset(n.heartbeatTimeout)
	}
	

	fmt.Printf("%s Will listen to heartbeat timer channel",n.id)
	
	go func() {
		select {
		case _,ok := <-n.heartbeatTimer.Channel():
			if ok {
			fmt.Printf("%s: got tick on heartbeat timer channel\n",n.id)
				n.eventCh <- new(TimeForHeartbeat)
			}
		}
	}()
}

func sendHeartbeat(n *node) {

	peers := n.config.Peers()
	// send heartbeat to all peers
	e := entry { from: n.id, term:n.currentTerm }
	fmt.Printf("%s sending hertbeat with term %d\n",n.id,e.term)

	fmt.Print("Peers: ")
	fmt.Println(peers)
	
	for _,p := range peers {
		if p.Id != n.id {

			fmt.Printf("%s - will send heartbeat event to %s\n",n.id,p.Id)
			
			go func(p Peer,e entry) {
				fmt.Printf("%s: sending heartbeat to peer : %s\n",n.id,p.Id)
				// send heartbeat
				reply,term := n.transport.AppendEntry(e,p)
				fmt.Printf("%s : got reply to append entry from peer: %s, %t\n",n.id,p.Id,reply)
				appendReply := AppendReply { reply:reply,term:term,from:p.Id }
				go func(ar AppendReply){
					fmt.Printf("%s - sending append entry response as event\n",n.id)
					n.eventCh <- &ar
				}(appendReply)
			}(p,e)
		}
	}
	fmt.Printf("%s - sent hertbeats\n",n.id)
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
		
		go func (p Peer,vreq voteRequest) {
			fmt.Printf("Requesting vote from: %s\n",p.Id)
			vres,err := n.transport.RequestForVote(vreq,p)
			if err == nil {
				fmt.Printf("Got vote from: %s\n",vres.from)
				//response got
				voteFrom := VoteFrom { from:vres.from, termToUpdate:vres.termToUpdate, voteGranted:vres.voteGranted }

				n.ifHigherTerm(vres.termToUpdate)
				
				fmt.Println("Pushing votefrom as event")
				n.eventCh <- &voteFrom
				fmt.Println("pushed votefrom as event")
			} else {
				// log the error?
			}
		}(p,vreq)
	}
}

