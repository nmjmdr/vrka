package timerwrap

import (
	"time"
	"fmt"
)

type TimerWrap interface {
	Channel() (<-chan time.Time)	
	Stop()
	Reset(d time.Duration) bool
}


type builtInTimer struct {
	t *time.Timer
}

func NewTimer(d time.Duration) TimerWrap {
	b := new(builtInTimer)
	fmt.Println("Called builtin timer")
	b.t = time.NewTimer(d)
	return b
}

func (b *builtInTimer) Channel() (<-chan time.Time) {
	return b.t.C
}

func (b *builtInTimer) Stop() {
	b.t.Stop()	
}

func (b *builtInTimer) Reset(d time.Duration) bool {
	return b.t.Reset(d)
}

type MockTimer struct {
	c chan time.Time
}

func NewMockTimer() TimerWrap {
	m := new(MockTimer)
	m.c = make(chan time.Time)
	return m
}


func (m *MockTimer) Channel() (<-chan time.Time) {
	return m.c
}

func (m *MockTimer) Stop() {
	close(m.c)
}

func (m *MockTimer) Tick() {
	fmt.Println("Mock timer - tick")
	m.c <- time.Time{}
}

func (m *MockTimer) Reset(d time.Duration) bool {
	return true
}




