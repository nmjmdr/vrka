package raft

import (
	"fmt"
	"timerwrap"
	"time"
	"math/rand"
	"sync"
)

type role int

const (
	Leader = 1
	Candidate = 2
	Follower = 3
)

const MinElectionTimeout = 150
const MaxElectionTimeout = 300

type voteRequest struct {
	term uint64
	candidateId string
	lastLogIndex uint64
	lastLogTerm uint64
}

type voteResponse struct {
	termToUpdate uint64
	voteGranted bool
	from string
}


type node struct {
	id string
	votesGot uint32
	outstandingVotes uint32

	currentTerm uint64
	role role
	leader Peer

	eventCh chan interface{}

	electionTimer timerwrap.TimerWrap

	config Config
	transport Transport

	getTimer getTimerFn

	electionTimeout time.Duration

	wg *sync.WaitGroup
}


type getTimerFn func(d time.Duration) timerwrap.TimerWrap


func newNode(id string,config Config,transport Transport,g getTimerFn) *node {
	n := new(node)
	n.id = id
	n.eventCh = make(chan interface{})
	n.currentTerm = 0

	n.config = config
	n.transport = transport

	n.getTimer = g

	n.electionTimeout = getRandomTimeout(MinElectionTimeout,MaxElectionTimeout)

	n.wg = &sync.WaitGroup{}

	n.loop()
	
	return n
}

func getRandomTimeout(startRange int,endRange int) time.Duration {
	timeout := startRange + rand.Intn(endRange - startRange)

	return time.Duration(timeout) * time.Millisecond
}


func (n *node) loop() {

	go func() {
		defer n.wg.Done()
		n.wg.Add(1)
		for  {
			fmt.Println("invoking select")
			select {
			case e,ok := <-n.eventCh:
				fmt.Printf("got event, and ok is: %d\n",ok)
				if ok {

					_,quit := e.(*Quit)
					if !quit {
						n.dispatch(e)
					} 
				}
				return
			}
		}		
	}()
}



func (n *node) dispatch(evt interface{}) {

	switch t := evt.(type) {
	default:
		panic(fmt.Sprintf("Unexpected event: %T",t))
	case *Initialized:
		// initialize
		// set the role as Follower
		fmt.Println("Init event received")
		n.role = Follower
		startElectionTimer(n)
		
		
	case *ElectionAnounced:
		// election anounced

		fmt.Println("Got election anounced event")
		
		if n.role == Follower {
			n.role = Candidate
			// start the election
		} else if n.role == Candidate {
			// did not get elected within the time, restart the election
		} else {
			panic("Got election anouncement while being a leader")
		}
		
	/*
	case *VoteFrom:
		// check if we got the vote
		majority := (n.config.Peers / 2) + 1
		if t.voteGranted {
			n.votesGot++
			if n.votesGot >= majority {
				// got elected
				r.role = Leader
			}
		}
*/
		
	}


}
