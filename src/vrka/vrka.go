package vrka

import (
	"howler"
	"time"
	"tickerwrap"
	"cbuckets"
)

type Vrka interface {
	Add(uri string,payload string,after time.Duration) (string,error)
	Del(string) error
}

type server struct {
	h howler.Howl
}

func New() Vrka {
	s := new(server)

	f := func(d time.Duration) tickerwrap.Tickerw {
		return tickerwrap.NewBuiltInTicker(d)
	}
	s.h = cbuckets.NewBuckets(f)
	return s
}

func (s *server) Add(uri string,payload string,after time.Duration) (string,error) {
	return "",nil
}

func (s *server) Del(id string) error {
	return nil
}

