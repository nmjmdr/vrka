package vrka

import (
	"common"
	"queue"
	"flake"
	"time"
	"tickerwrap"
	"fmt"
)

const qFactor = 2

type CallbackServer interface {
	Add(afterms uint64,uri string) string
	Channel() <-chan (*common.Reminder)
	Start()
	Stop()
}

type qServer struct {
	flk *flake.Flk
	q queue.Queue
	stop chan bool
	remc chan (*common.Reminder)
	tw tickerwrap.Tickerw
}

type CreateTicker func(d time.Duration) tickerwrap.Tickerw

func NewServer(c CreateTicker,tickerInterval time.Duration) CallbackServer {
	s := new(qServer)
	var err error
	s.flk,err = flake.FlakeNode()
	if err != nil {
		panic(err)
	}
	s.q = queue.NewQueue(qFactor)
	s.stop = make(chan bool)
	s.remc = make(chan (*common.Reminder))
	s.tw = c(tickerInterval)
	return s
}

func (s *qServer) Add(afterms uint64,uri string) string {
	idBytes,err := s.flk.NextId()
	if err != nil {
		panic(err)
	}

	id := string(idBytes)
	r := new(common.Reminder)
	r.Type = common.Callback
	r.Id = id
	r.Payload = uri
	// ms to nanoseconds
	nano := nanoTimeStamp(afterms)
	// nano - the value of UniNano after "x" ms from
	// current UnixNamo value
	s.q.Add(nano,r)
	return id
}

func nanoTimeStamp(ms uint64) uint64 {
	//ms := nano / (int64(time.Millisecond)/int64(time.Nanosecond))
	nanoAfter :=  ms * uint64(int64(time.Millisecond)/int64(time.Nanosecond))

	fmt.Printf("at : %d\n",uint64(time.Now().UnixNano()) + nanoAfter)
	return uint64(time.Now().UnixNano()) + nanoAfter
}



func (s *qServer) Channel() <-chan (*common.Reminder) {
	return s.remc
}

func (s *qServer) tickHandler() {

	// need to wait on the ticker channel
	for range s.tw.Channel() {
		r,f := s.q.GetMinInRange(0,uint64(time.Now().UnixNano()))
		if f {
			fmt.Printf("got: %s",r.Payload)
			s.remc <- r
		}
        }
}

func (s *qServer) Start() {
	go s.tickHandler()
}


func (s *qServer) Stop() {
	s.tw.Stop()
}
