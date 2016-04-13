package raft

import (
	"testing"
	"timerwrap"
	"time"
	"sync"
)


func Test_Timeout(t *testing.T) {


	tw := timerwrap.NewMockTimer()
	
	f := func(d time.Duration) timerwrap.TimerWrap  {
		return tw
	}
	
	m := NewMonitor(f)
	m.Start()

	gotNotice := false	

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		select {
		case <-m.ElectionNotice():
			gotNotice = true			
		}
		wg.Done()
	}()

	mockTimer,_ := tw.(*timerwrap.MockTimer)
	mockTimer.Tick()

	wg.Wait()
	if !gotNotice {
		t.Fail()
	}

}


func Test_NoElectionNotice(t *testing.T) {

	tw := timerwrap.NewMockTimer()
	
	f := func(d time.Duration) timerwrap.TimerWrap  {
		return tw
	}
	
	m := NewMonitor(f)
	m.Start()

	gotNotice := false	

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {

		mInternal,_ := m.(*monitor)
		select {
		case <-m.ElectionNotice():
			gotNotice = true
		case <-mInternal.heartbeat:
		}
		wg.Done()
	}()

	m.Heartbeat()	
	mockTimer,_ := tw.(*timerwrap.MockTimer)
	mockTimer.Tick()	
	
	wg.Wait()
	if gotNotice {
		t.Fail()
	}
}
