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
	termToUpdate uint64
	voteGranted bool
}


type candidate struct {
	votes int
	mutex *sync.Mutex
	peerResponse chan VoteResponse
	r *raftNode
	noticeCh chan bool
	quitCh chan bool
}


func NewCandidate(r *raftNode) state {
	c := new(candidate)
	c.r = r
	c.mutex = &sync.Mutex{}
	c.peerResponse = make(chan VoteResponse)
	c.noticeCh = make(chan bool)
	c.quitCh = make(chan bool)
	
	return c
}

func (c *candidate) askForVotes() {

	

	// vote for self
	c.votes = c.votes + 1

	// three scnearios:
	// 1. Majority votes received, elected as leader
	// 2. Received another Append entries (heartbeat) with higher or equal term, becomea follower
	// 3. Election timeout - nobody elected as leader
	
	
}

func (c *candidate) sendRequests() {

/* when sending requests there are these cases:
1. The node obtains majority of votes -> transitions to a leader
2. The node does not obtain a majority of votes -> other followers have already voted, but yet to receive heartbeats
      - the candidate will eventually receive the heartbeat and become a follower
3. The node does not obtain a majority of votes -> other followers have already voted, have already received a heartbeat
      - the candidate will eventually receive the heartbeat and become a follower
4. Nobody get elected, votes split
      - will timeout on  election notice
5. No responses obtained timeout, may be the network was broken midway, the candidate got disconnected from network
     - timeout on election notice
*/

	
	
	vreq := VoteRequest{}
	vreq.term = c.r.currentTerm
	vreq.candidateId = c.r.id
	vreq.lastLogIndex = 0 // set this later
	vreq.lastLogTerm = 0 // set this later
	
	
	peers := c.r.config.Peers()	
	
	wg := sync.WaitGroup{}
	// expect for the current node
	wg.Add(len(peers)-1)

	for i:=0;i<len(peers);i++ {
		if peers[i].Id == c.r.id {
			continue
		}

		go func() {
			// try and optimize this by making the network calls async??
			vres,err := c.r.transport.RequestForVote(vreq)
			if err != nil && vres.voteGranted {
				c.mutex.Lock()
				c.votes++
				c.mutex.Unlock()
			}
			wg.Done()			
		}()
	}


	// signal wait done
	fin := make(chan bool)
	go func() {
		defer close(fin)
		wg.Wait()
	}()


	select {
	case <- c.noticeCh:
		// nobody got elected
	case <- fin:
		// check if got the requiste number of votes
		if c.votes >= (len(peers)/2 + 1) {
			// transition to leader
			c.r.mutex.Lock()
			c.r.role = Leader
			c.r.state = NewLeader(c.r)
			c.r.mutex.Unlock()
		} else {
			// just wait for the time out notice and then start the election afresh????
		}
		
	}
}

func (c *candidate) onElectionNotice() state {
	// notice to stop asking for votes
	c.noticeCh <- true
	
	return c
}

func (c *candidate) onQuit() state {
	return c
}


func (c *candidate) onHeartbeat(beat Beat) state {

	
	//change role to follower, if the term is equal to or greater
	// than the candiate's term
	c.r.mutex.Lock()

	if c.r.currentTerm <= beat.Term {
		c.r.role = Follower
		c.r.monitor.Reset()
	}	
	c.r.mutex.Unlock()

	// stop the election notices, reset the election timer
	f := NewFollower(c.r)
	return f
}
