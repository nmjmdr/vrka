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

const currentTermKey = "Current_Term"
const votedForKey = "Voted_For_Key"


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

type entry struct {
	term uint64
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
	stable Stable

	getTimer getTimerFn

	electionTimeout time.Duration

	wg *sync.WaitGroup

	stateChange chan role
}


type getTimerFn func(d time.Duration,nodeId string) timerwrap.TimerWrap


func newNode(id string,config Config,transport Transport,g getTimerFn,stable Stable) *node {
	n := new(node)
	n.id = id
	n.eventCh = make(chan interface{})
	n.currentTerm = 0

	n.config = config
	n.transport = transport
	n.stable = stable

	n.getTimer = g

	n.electionTimeout = getRandomTimeout(MinElectionTimeout,MaxElectionTimeout)

	n.wg = &sync.WaitGroup{}

	n.stateChange = make(chan role)

	n.role = Follower

	//set term from stable
	n.setCurrentTermFromStable()

	n.loop()
	
	return n
}

func (n *node) setCurrentTermFromStable() {
	term,ok := n.stable.GetUint64(currentTermKey)
	if !ok {
		// probably this is the first time we are running this node
		n.currentTerm = 0
		n.stable.Store(currentTermKey,n.currentTerm)
		return
	}
	n.currentTerm = term
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
				if ok {
					fmt.Printf("%s: got event\n",n.id)
					_,quit := e.(*Quit)
					if !quit {
						n.dispatch(e)
					} else {
						fmt.Printf("%s node quiting...\n",n.id)
						n.stateChange <- n.role
						return
					}
				}
			}
		}		
	}()
}


func (n *node) startAsFollower() {
	// initialize
	// set the role as Follower
	n.role = Follower
	fmt.Printf("%s: set as follower\n",n.id)
	n.anounceRoleChange()
	fmt.Println("Start Follower event received")
	startElectionTimer(n)
}

func (n *node) restartElection() {
	n.anounceRoleChange()
	// start the election
	// start Election timer
	startElectionTimer(n)
	startElection(n)
}


func (n *node) dispatch(evt interface{}) {
	
	switch t := evt.(type) {
	default:
		panic(fmt.Sprintf("Unexpected event: %T",t))
	case *StartFollower:
		{
			n.startAsFollower()
		}
		
	case *ElectionAnounced:
		{
	
			fmt.Println("Got election anounced event")		
			if n.role == Follower {
				fmt.Println("Changing to candidate and starting election")
				n.role = Candidate
				n.restartElection()
			} else if n.role == Candidate {
				// did not get elected within the time, restart the election
				fmt.Println("did not get elected, starting election again")
				n.restartElection()
			} else {
				panic("Got election anouncement while being a leader")
			}

		}
	
	case *VoteFrom:
		{
			// check if we got the vote
			if n.role == Follower {
				// reject this, might have been a delayed response
				// from a node
			} else if n.role == Candidate {
				n.handleVoteFrom(t)
			} else if n.role == Leader {
				// delayed vote ignore
			}
		}

	case *HigherTermDiscovered:
		{
			if t.term > n.currentTerm {
				fmt.Printf("%s: Higher term discovered\n",n.id)
				// revert to follower
				n.currentTerm = t.term
				n.startAsFollower()			
			}
		}
	}
}

func (n *node) anounceRoleChange() {
	fmt.Printf("%s : Anouncing state change to: %d\n",n.id,n.role)
	n.stateChange <- n.role
}

func (n *node) handleVoteFrom(v *VoteFrom) {

	fmt.Printf("%s: got vote from: %s\n",n.id,v.from)
	if v.voteGranted {
		majority := uint32((len(n.config.Peers()) / 2) + 1)
		n.votesGot++
		fmt.Printf("%s : votes got: %d\n",n.id,n.votesGot)
		if n.votesGot >= majority {
			// got elected
			fmt.Println("setting as leader")
			n.role = Leader
			n.anounceRoleChange()
		}
	}
	// else - what do we do??
	
	
}
