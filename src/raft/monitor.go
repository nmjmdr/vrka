package raft

import (
	"time"
	"math/rand"
)

const MinTimeout = 150
const MaxTimeout = 300

type Monitor interface {	
	Stop() bool
	ElectionNotice() (<-chan time.Time)
	Reset()
}

type monitor struct {
	timeout uint32
	timer *time.Timer
	
}

func NewMonitor() Monitor {
	m := new(monitor)
	m.timeout = MinTimeout + uint32(rand.Intn(MaxTimeout - MinTimeout))
	m.timer = time.NewTimer(time.Duration(m.timeout) * time.Millisecond)
	return m	
}


func (m *monitor) Stop() bool {
	return m.timer.Stop()
	
}

func (m *monitor) ElectionNotice() (<-chan time.Time) {
	return m.timer.C
}

func (m *monitor) Reset() {
	m.timer.Reset(time.Duration(m.timeout) * time.Millisecond)	
}
