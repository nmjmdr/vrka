package raft

import (
	"time"
	"math/rand"
	"timerwrap"
)

type Monitor interface {
	Start()
	Stop()
	Heartbeat()
	ElectionNotice() (<-chan bool)
}

const minTimeout = 150
const maxTimeout = 300

type CreateTimer func(d time.Duration) timerwrap.TimerWrap

type monitor struct {
	timeout uint32
	heartbeat chan bool	  
	timer timerwrap.TimerWrap
	election chan bool
	quit chan bool
	createTimer CreateTimer
}



func NewMonitor(createTimer CreateTimer) Monitor {

	m := new(monitor)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m.timeout = uint32((r.Float64() * (maxTimeout-minTimeout) + minTimeout))
	m.heartbeat = make(chan bool)
	m.election = make(chan bool)
	m.quit = make(chan bool)
	m.createTimer = createTimer
	return m
}

func (m *monitor) Heartbeat() {
	// every heartbeat received resets the ticker
	m.heartbeat <- true
}

func (m *monitor) ElectionNotice() (<-chan bool) {
	return m.election
}

func (m *monitor) Start() {
	m.timer = m.createTimer(time.Duration(m.timeout) * time.Millisecond)
	// start monitoring for timeout
	go func() {
		for {
			select {
			case <-m.timer.Channel():
				//timed out start - become a candiate
				m.election <- true
			case <-m.heartbeat:
			// reset the timer
				m.timer.Reset(time.Duration(m.timeout) * time.Millisecond)
			case <-m.quit:
				m.timer.Stop()
				break
			}
		}
	}()
}


func (m *monitor) Stop() {
	m.quit <- true
}
