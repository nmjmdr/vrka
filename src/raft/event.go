package raft

type EvtType int

const (
	HeartbeatEvtType = 1
	ElectionTimeoutEvtType = 2
	ElectionResultEvtType = 3
	QuitEvtType = 4	
)


type Evt interface {
	Type() EvtType
}

type heartbeatEvt struct {
	e entry
}

func (h *heartbeatEvt) Type() EvtType {
	return HeartbeatEvtType
}

type electionTimoutEvt struct {
}

func (e *electionTimoutEvt) Type() EvtType {
	return ElectionTimeoutEvtType
}

type quitEvt struct {
}


func (q *quitEvt) Type() EvtType {
	return QuitEvtType
}


type electionResultEvt struct {
	elected bool
}

func (e *electionResultEvt) Type() EvtType {
	return ElectionResultEvtType
}

