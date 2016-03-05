package vrka

import (
	"howler"
	"time"
)

type Vrka interface {
	Add(uri string,payload string,after time.Duration) (string,error)
	Del(string) error
}

type server struct {
	h *howler.Howl
}

func (s *server) Add(uri string,payload string,after time.Duration) (string,error) {
	return "",nil
}
func Del(id string) error {
	return nil
}

