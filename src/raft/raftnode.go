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


type beat struct {
	term uint64
	from string
}


type RaftNode interface {
	Heartbeat(beat beat)
	Stop()
	CurrentRole() role
	RoleChange() <- chan role
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
	heartbeatCh chan beat
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
	r.heartbeatCh = make(chan beat)
	r.roleChangeCh = make(chan role)
	r.electionResultCh = make(chan bool)

	r.monitor = monitor
	r.wg = &sync.WaitGroup{}

	r.stateFn = followerFn

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

func (n *raftNode) Heartbeat(beat beat) {
	n.heartbeatCh <- beat
}

func (n *raftNode) CurrentRole() role {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.role
}

func (n *raftNode) anounceRoleChange(role role) {
	select {
	case n.roleChangeCh <- role:
	default:
	}
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
					fmt.Printf("role: %s\n",node.id)					
					node.stateFn(node,&(electionTimoutEvt{}))
					node.anounceRoleChange(node.role)
				}

			case beat,ok := <-node.heartbeatCh:
				if ok {
					hb := new(heartbeatEvt)
					hb.beat = beat
					node.stateFn(node,hb)
					node.anounceRoleChange(node.role)				
				}

			case result,ok := <-node.electionResultCh:
				if ok {
					fmt.Println("got election result")
					resultEvt := new(electionResultEvt)
					resultEvt.elected = result
					node.stateFn(node,resultEvt)
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

