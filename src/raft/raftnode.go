package raft

import (
	"fmt"
	"sync"
)

type role int

const (
	leader = 1
	candidate = 2
	follower = 3
)


type entry struct {
	term uint64
	leaderId string
	prevLogIndex uint64
	prevLogTerm uint64
	entries []byte
	leaderCommit uint64
}


type RaftNode interface {
	Append(entry entry)
	Stop()
	CurrentRole() role
	RoleChange() <- chan role
	RequestForVote(vreq voteRequest) voteResponse
}

type raftNode struct {
	id string
	role role
	leader Peer
	votedFor string
	config Config
	term uint64
	transport Transport

	mutex *sync.Mutex
	monitor Monitor

	quitCh chan bool
	appendCh chan entry
	roleChangeCh chan role
	electionResultCh chan bool

	wg *sync.WaitGroup

	stateFn stateFunction
}

type stateFunction func(r *raftNode,evt Evt)

func NewRaftNode(id string,monitor Monitor,config Config,transport Transport) RaftNode {
	r := new(raftNode)
	r.id = id
	r.role = follower
	r.mutex = &sync.Mutex{}
	r.config = config
	r.transport = transport
	r.term = 0

	r.quitCh = make(chan bool)
	r.appendCh = make(chan entry)
	r.roleChangeCh = make(chan role)
	r.electionResultCh = make(chan bool)

	r.monitor = monitor
	r.wg = &sync.WaitGroup{}

	r.stateFn = followerFn

	r.votedFor = ""
	
	loop(r)

	return r
}


func (n *raftNode) RoleChange() <- chan role {
	return n.roleChangeCh
}

func (n *raftNode) Stop() {
	n.quitCh <- true
	n.wg.Wait()
}

func (n *raftNode) Append(e entry) {
	n.appendCh <- e
}

func (n *raftNode) CurrentRole() role {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.role
}

func (n *raftNode) anounceRoleChange(role role) {
	select {
	case n.roleChangeCh <- role:
		//default:
		// un comment this for real run
		// udpate this section later
	}
}

func (r *raftNode) RequestForVote(vreq voteRequest) voteResponse {

	vres := voteResponse{}

	defer r.mutex.Unlock()
	r.mutex.Lock()
		
	if vreq.term < r.term {
		vres.voteGranted = false
		vres.termToUpdate = r.term
		return vres
	}

	
	if len(r.votedFor) == 0 {
		// need to extend this condition: If votedFor is null or candidateId, and candidate's log is at least as up-to-date as receiver's log, grant vote
		vres.voteGranted = true
		// set the term and voted for
		r.votedFor = vreq.candidateId
		r.term = vreq.term

		return vres		
	}

	vres.termToUpdate = r.term
	vres.voteGranted = false

	return vres	
}

func loop(node *raftNode) {	
	
	go func() {
		node.wg.Add(1)
		defer node.wg.Done()
 		quit := false
			
		for !quit {
			fmt.Println("will wait on select")
			select {
			case _,ok := <-node.monitor.ElectionNotice():	
				if ok {
					fmt.Println("got election notice")
					fmt.Printf("node id: %s\n",node.id)					
					node.stateFn(node,&(electionTimoutEvt{}))
					fmt.Printf("will anounce role change: %d\n",node.role)
					node.anounceRoleChange(node.role)
				}

			case e,ok := <-node.appendCh:
				if ok {
					if e.entries == nil || len(e.entries) == 0 {
						hb := new(heartbeatEvt)
						hb.e = e
						node.stateFn(node,hb)
						fmt.Printf("will anounce role change: %d\n",node.role)
						node.anounceRoleChange(node.role)
					}
				}

			case result,ok := <-node.electionResultCh:
				if ok {
					fmt.Println("got election result")
					resultEvt := new(electionResultEvt)
					resultEvt.elected = result
					node.stateFn(node,resultEvt)
					fmt.Printf("will anounce role change: %d\n",node.role)
					node.anounceRoleChange(node.role)
				}

			

			case <-node.quitCh:
				fmt.Println("got quit")
				quit = true
				break
			}
		}
	}()
}

