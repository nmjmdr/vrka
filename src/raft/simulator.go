package raft

import (
	"time"
	"strconv"
	"timerwrap"
)

type Simulator struct {
	StateChanges map[string](chan string)
	quitCh chan bool
	nodes [](*node)
	numNodes int
}


func NewSimluator(numNodes int) *Simulator {

	s := new(Simulator)

	s.StateChanges = make(map[string](chan string))
	s.numNodes = numNodes
	s.quitCh = make(chan bool)

	return s		
}

func (s* Simulator) Stop() {
	s.quitCh <- true
}

func (s *Simulator) Start() []string {

	peers := make([]Peer,s.numNodes)
	transport := newInMemoryTransport()

	nodeIds := make([]string,s.numNodes)
	
	for i:=0;i<s.numNodes;i++ {
		str := strconv.Itoa(i)
		peers[i] =  Peer{Id:str,Address:""}	
	}

	s.nodes = make([](*node),s.numNodes)

	for i:=0;i<s.numNodes;i++ {
		str := strconv.Itoa(i)
		nodeIds[i] = str
		s.nodes[i] = makeNewNode(str,peers,transport)
		transport.setNode(str,s.nodes[i])
		s.StateChanges[str] = make(chan string)
		go s.listenToStateChange(s.nodes[i])
		start(s.nodes[i])
	}

	return nodeIds
}

func (s *Simulator) listenToStateChange(n *node) {

	for  {
		select {
		case role, ok := <- n.stateChange:
		{
			if ok {
				s.StateChanges[n.id] <- roleInfo(role)
			}
		}
		case _,_ = <- s.quitCh:
			// stop all
			for _,node := range s.nodes {
				stop(node)
			}
			return
		}
	}	
}

func roleInfo(r role) string {
	switch r {
	case Follower:
		return "Follower"
	case Leader:
		return "Leader"
	case Candidate:
		return "Candidate"

	default:
		return "Unknown"
	}
	
}



func makeNewNode(id string,peers []Peer,transport Transport) *node {

	g := func(d time.Duration,nodeId string,forHint string) timerwrap.TimerWrap {
		return timerwrap.NewTimer(d)
	}	

	config := NewMockConfig(peers)
	stable := newInMemoryStable()

	return newNode(id,config,transport,g,stable)
}
