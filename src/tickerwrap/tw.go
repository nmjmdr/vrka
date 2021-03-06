

package tickerwrap

import (
	"time"
)

type Tickerw interface {
	Start()
	Channel() (<-chan time.Time)
	Stop()
}

type BuiltInTicker struct {
	t *time.Ticker
	d time.Duration
}


func NewBuiltInTicker(d time.Duration) Tickerw {
	b := new(BuiltInTicker)
	b.d = d
	return b
}



func (b *BuiltInTicker) Start() {
	b.t = time.NewTicker(b.d)
}

func (b *BuiltInTicker) Channel() (<-chan time.Time) {
	return b.t.C
}

func (b *BuiltInTicker) Stop() {
	if b.t != nil {
		b.t.Stop()
	}
}


type MockTicker struct {
	c (chan time.Time)
}

func NewMockTicker() Tickerw {
	m := new(MockTicker)
	m.c = make(chan time.Time)
	return m
}

func (m *MockTicker) Start() {
}

func (m *MockTicker) Channel() (<-chan time.Time) {
	return m.c
}

func (m *MockTicker) Stop() {
}

func (m *MockTicker) Tick() {
	m.c <- time.Time{}
}

