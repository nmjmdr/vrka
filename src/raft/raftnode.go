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

const HeartbeatTimeout = 50

const ForElectionTimer = "ForElectionTimer"
const ForHeartbeatTimer = "ForHeartbeatTimer"


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
	heartbeatTimer timerwrap.TimerWrap

	config Config
	transport Transport
	stable Stable

	getTimer getTimerFn

	electionTimeout time.Duration
	heartbeatTimeout time.Duration

	wg *sync.WaitGroup

	stateChange chan role
}


type getTimerFn func(d time.Duration,nodeId string,forHint string) timerwrap.TimerWrap


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
	n.heartbeatTimeout = time.Duration(HeartbeatTimeout * time.Millisecond)

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
	// are we transitioning from a leader role?

	if n.role == Leader {
		fmt.Printf("%s - Will stop heartbeat timer\n",n.id)
		stopHeartbeatTimer(n)
	}
	
	n.role = Follower
	fmt.Printf("%s: set as follower\n",n.id)
	startElectionTimer(n)
	n.anounceRoleChange()
}

func (n *node) restartElection() {	
	// start the election
	// start Election timer
	startElectionTimer(n)
	startElection(n)
	n.anounceRoleChange()
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
	
			fmt.Printf("%s - Got election anounced event\n",n.id)		
			if n.role == Follower {
				fmt.Printf("%s - Changing to candidate and starting election\n",n.id)
				n.currentTerm = n.currentTerm + 1
				n.role = Candidate
				n.restartElection()
			} else if n.role == Candidate {
				// did not get elected within the time, restart the election
				fmt.Printf("%s - did not get elected, starting election again\n",n.id)
				n.currentTerm = n.currentTerm + 1
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
				n.handleHigherTermDiscovered(t.term)		
			}
		}

	case *Append:
		{
			if t.term > n.currentTerm {
				n.handleHigherTermDiscovered(t.term)		
			} else {
				// we reset the election timer
				fmt.Printf("%s - will reset election timer to: ",n.id)
				fmt.Println(n.electionTimeout)
				n.electionTimer.Reset(n.electionTimeout)
				n.currentTerm = t.term
			}
		}
	case *AppendReply:
		{
			fmt.Printf("%s - Append reply as event, with term : %d, current term is: %d\n",n.id,t.term,n.currentTerm)
			if t.term > n.currentTerm {
				n.handleHigherTermDiscovered(t.term)		
			}
			// if not perform other aspects of append reply
		}

	case *TimeForHeartbeat:
		{			
			if n.role == Leader {
				fmt.Printf("%s node - time for heartbeat, will send out heartbeat\n",n.id)
				sendHeartbeat(n)
				startHeartbeatTimer(n)				
			} else {
				// good chance that the timer is stopped
				// but we get the evet from the previous timer expiry
				// ignore it
				fmt.Printf("%s Heartbeat timer is active while the node is not a leader - ignoring it\n",n.id)
				return
			}
			
		}
	}
}

func (n *node) handleHigherTermDiscovered(termToUpdate uint64) {
	fmt.Printf("%s: Higher term discovered\n",n.id)
	// revert to follower
	n.currentTerm = termToUpdate
	n.startAsFollower()
}

func (n *node) anounceRoleChange() {
	fmt.Printf("%s : Anouncing state change to: %d\n",n.id,n.role)
	n.stateChange <- n.role
}

func (n *node) startAsLeader() {
	fmt.Printf("%s - setting as leader",n.id)
	n.role = Leader
	// stop election timer
	stopElectionTimer(n)
	// start leader heartbeat
	sendHeartbeat(n)
	startHeartbeatTimer(n)
	n.anounceRoleChange()
}

func (n *node) handleVoteFrom(v *VoteFrom) {

	fmt.Printf("%s: got vote from: %s\n",n.id,v.from)
	if v.voteGranted {
		majority := uint32((len(n.config.Peers()) / 2) + 1)
		n.votesGot++
		fmt.Printf("%s : votes got: %d\n",n.id,n.votesGot)
		if n.votesGot >= majority {
			// got elected
			n.startAsLeader()
		}
	}
	// else
	// if a higher term has been discovered the event higher term discovered
	// would already be placed 
	
	
}
