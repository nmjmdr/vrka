package vrka

import (
	"howler"
	"time"
	"tickerwrap"
	"cbuckets"
	"fmt"
)

type Vrka interface {
	Add(uri string,payload string,afterms uint64 ) (string,error)
	Del(string) error
	Stop()
}

type server struct {
	h howler.Howl
	quit chan bool
	cbchan chan howler.Callback
}

func New() Vrka {
	s := new(server)

	f := func(d time.Duration) tickerwrap.Tickerw {
		return tickerwrap.NewBuiltInTicker(d)
	}
	s.h = cbuckets.NewBuckets(f)

	go s.handleCallback(s.h)
	

	return s
}

func (s *server) handleCallback(h howler.Howl) {
	for {
		select {
		  case cb := <-h.Howls():
			makeCallback(cb)
		  case _ = <-s.quit:
			break
		}
	}
}


func makeCallback(cb howler.Callback) {
	// make the callback 
	// invoke the uri
	// if it fails we could it as a callback 
	// again, but set a counter to indicate 
	// that the callback was attempted
	fmt.Println(cb.Uri)
}

func (s *server) Stop() {
	s.quit <- true
}

func (s *server) Add(uri string,payload string,afterms uint64) (string,error) {

	var c  howler.Callback
	c.Uri = uri
	c.Payload = payload

	id, err := s.h.Add(c,afterms)

	return id,err
}

func (s *server) Del(id string) error {

	err := s.h.Del(id)
	return err
}

