package vrka

import (	
	"net/http"
	"fmt"
	"time"
)


type Caller interface {
	Call(uri string)
}

const timeoutMs = 25

type httpCaller struct {
	okCh chan bool
	errCh chan error
}

func NewCaller() Caller {
	h := new(httpCaller)
	h.okCh = make(chan bool)
	h.errCh = make(chan error)
	return new(httpCaller)
}

// Probably a need a better way to perform these operations
// should not be waiting for 25 ms for a timeout
// it possible to do a one-way call kind of semantics
// end the momemt the first byte is received?
func (h *httpCaller) Call(uri string) {

	// make the call and wait for
	// response or a timeout
	// upon timeout we can insert the request to a failed queue
	
	go func() {
		// print diagnostic message
		fmt.Printf("get: %s - ",uri)
		resp, err := http.Get(uri)
		resp.Body.Close()
		if err != nil {
			h.errCh <- err
		} else {
			h.okCh <- true
		}
	}()

	select {
	case <-h.okCh:
		fmt.Println("ok")
	case e:=<-h.errCh:
		fmt.Println(e)
	case <-time.After(timeoutMs * time.Millisecond):
		fmt.Println("timed out")
	}
}
