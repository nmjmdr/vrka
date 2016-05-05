package raft

import (
	"time"
)

type mockTransport struct {
	vres voteResponse
	err error
	delay bool
	sleepTime time.Duration
}


func NewMockTransport(vres voteResponse) *mockTransport {
	m := new(mockTransport)
	m.vres = vres

	return m
}

func NewMockTransportWithDelay(vres voteResponse,delayTime time.Duration) *mockTransport {
	m := new(mockTransport)
	m.vres = vres
	m.delay = true
	m.sleepTime = delayTime
	return m
}

func NewMockTransportError(err error) *mockTransport {
	m := new(mockTransport)
	m.err = err
	return m
}

func (m *mockTransport) RequestForVote(vreq voteRequest) (voteResponse,error) {
	if m.delay {
		time.Sleep(m.sleepTime)
	}
	return m.vres,m.err
}
