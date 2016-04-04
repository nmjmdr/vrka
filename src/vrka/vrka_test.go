package vrka

import (
	"testing"	
	"tickerwrap"
	"time"
	"fmt"	
)



func Test_SimpleAdd(t *testing.T) {
	
	d := 1 * time.Millisecond

	ticker := tickerwrap.NewMockTicker()
	f := func(d time.Duration) tickerwrap.Tickerw {
		return ticker
	}
	
	c := NewServer(f,d)

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(2 * time.Millisecond)
		timeout <- true
	}()

		
	// wait for callback
	go func() {
		select {
		case r:=<-c.Channel():
			t.Log(r.Payload)			
		case <-timeout:
			t.Fail()
		}
	}()
	c.Start()
	c.Add(1,"uri")
	mockTicker,_ := ticker.(*tickerwrap.MockTicker)
	nano := msToNano(1)

	fmt.Printf("now: %d, nano: %d",time.Now().UnixNano(),nano)
	
	for i:=uint64(0);i<nano;i++ {
		mockTicker.Tick()
	}

	c.Stop()
	
	
}


func msToNano(ms uint64) uint64 {
	return  ms * uint64(int64(time.Millisecond)/int64(time.Nanosecond))
}
