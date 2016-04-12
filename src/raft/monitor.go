package raft

import (
	"time"
	"math/rand"
)

type Monitor interface {
	Start()
	Stop()
	Heartbeat()
}

const minTimeout = 150
const maxTimeout = 300

type monitor struct {
	timeout uint32
	heartbeat chan bool	  
	timer *time.Timer
	election chan bool
}

func NewMonitor() Monitor {

	m := new(monitor)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m.timeout = uint32((r.Float64() * (maxTimeout-minTimeout) + minTimeout))
	m.heartbeat = make(chan bool)
	m.election = make(chan bool)
	return m
}

func (m *monitor) Heartbeat() {
	// every heartbeat received resets the ticker
}


func (m *monitor) Start() {
	m.timer = time.NewTimer(time.Duration(m.timeout) * time.Millisecond)
	// start monitoring for timeout
	go func() {
		for {
			select {
			case <-m.timer.C:
				//timed out start - become a candiate
				m.election <- true
			case <-m.heartbeat:
			// reset the timer
				m.timer.Reset(time.Duration(m.timeout) * time.Millisecond)
			}
		}
	}()
}


func (m *monitor) Stop() {
}
