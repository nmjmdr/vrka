package howler

import (
	"time"
)

type Callback struct {
	Uri     string
	Payload string
}

type Howl interface {
	Add(Callback, time.Duration) (string, error)
	Del(string) error
	Howls() <-chan Callback
}
