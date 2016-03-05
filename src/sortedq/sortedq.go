package sortedq

import (
	"howler"
	"time"
)



// uses a skiplist to maintain a sorted queue
type Queue struct {
	c    (<-chan howler.Callback)
}

func NewSortedQ() *Queue {
	q := new(Queue)
	return q
}

func (q *Queue) Add(c howler.Callback, after time.Duration) (string, error) {
	return "", nil
}

func (q *Queue) Del(id string) error {
	return nil
}

func (q *Queue) Howls() <-chan howler.Callback {
	return q.c
}
