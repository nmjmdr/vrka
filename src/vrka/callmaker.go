package vrka

import (	
	"net/http"
	"fmt"
	"time"	
)


type Caller interface {
	Call(uri string)
}

const timeoutMs = 10000

type httpCaller struct {	
}

func NewCaller() Caller {
	return new(httpCaller)
}

// Probably a need a better way to perform these operations
// should not be waiting for "x" ms for a timeout
// it possible to do a one-way call kind of semantics
// end the momemt the first byte is received?
func (h *httpCaller) Call(uri string) {

	// make the call and wait for
	// response or a timeout
	// upon timeout we can insert the request to a failed queue
	timeoutCh := make(chan bool)
	errCh := make(chan error)
	okCh := make(chan bool)
	
	go func() {
		// print diagnostic message		
		fmt.Printf("GET : %s - ",uri)
		
		resp, err := http.Get(uri)		
		if err != nil {
			errCh <- err
		} else {		
			okCh <- true
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	go func() {
		time.Sleep(timeoutMs * time.Millisecond)
		timeoutCh <- true
	}()

	select {
	case <-okCh:
		fmt.Println("ok")
	case e:=<-errCh:
		fmt.Println(e)	
	case <-timeoutCh:
		fmt.Println("timed out")
	}
}
