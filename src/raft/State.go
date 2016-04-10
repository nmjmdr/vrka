package raft


type State struct {

	currentTerm uint64
	votedFor string
	
	commitIndex uint64
	lastApplied uint64
}

type LogState struct {
	nextIndex uint32
	matchIndex uint32
}


type LeaderState struct {
	// nextIndex and matchIndex
	// for each of the peers
	logState [] LogState
}
