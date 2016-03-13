package cbuckets

import (
	"howler"	
	"sync"
	"flake"
	"errors"
	"fmt"
	"time"
)

const numBuckets = 8
const factor = 16
const firstBucketStart = 0
const firstBucketEnd = 256


type node struct {
	id string
	c *howler.Callback
	after uint64 //in milliseconds
	next *node
	prev *node
	bucketIndex int
}

type bucket struct {
	start uint64
	end uint64
	head *node
	tail *node
	mutex *sync.Mutex	
}


type TimedBuckets struct {
	c  (<-chan howler.Callback)
	buckets []bucket
	m map[string](*node)
	flk *flake.Flk
	mutex *sync.Mutex
}

func (b *TimedBuckets) setup() {
	b.buckets = make([]bucket,numBuckets)

	start := uint64(firstBucketStart)
	end := uint64(firstBucketEnd)

	b.buckets[0].start = start
	b.buckets[0].end = end
	b.buckets[0].mutex = &sync.Mutex{}
	
	for n:=1;n<numBuckets;n++ {

		start = end+1
		end = end * factor

		b.buckets[n].start = start
		b.buckets[n].end = end	
		b.buckets[n].mutex = &sync.Mutex{}
	}
}


func (b *TimedBuckets) newId() string {
	id,err := b.flk.NextId()
	if err != nil {
		panic(err)
	}
	return string(id)
}

func NewBuckets() *TimedBuckets {
	b := new(TimedBuckets)
	b.setup()
	var err error
	b.flk,err = flake.FlakeNode()
	b.m = make(map[string](*node))
	b.mutex = &sync.Mutex{}
	if err != nil {
		panic(err)
	}
	return b
}

func (b *TimedBuckets) newNode(c *howler.Callback,after uint64) *node {

	id := b.newId()
	n := new(node)
	n.id = id
	n.c = c
	n.after = after	
	return n
}

func (b *TimedBuckets) addToBuckets(n *node) {

	
	i := 0
	for ;i<len(b.buckets);i++ {
		if b.buckets[i].start <= n.after && b.buckets[i].end >= n.after {
			break
		
		}
	}

	// set the node's bucket index
	n.bucketIndex = i

	// add to the tail, update the tail
	// now the conflict that could happen are:
	// the current tail is getting deleted
	// another rountine is trying to add another element
	// at the tail
	// the tail element is getting moved up a level
	// we have added the new element, by setting "next", but
	// before we could set the tail - the timer is trying to move
	// the element up one level


	// to avoid all this we take the lock
	// at the bucket level 
	// before inserting a node into it
	// or before deleting a node from it
	// or before moving a node one level up
	
	// check if it is nill
	// try and lock bucket 
	b.buckets[i].mutex.Lock()
	if b.buckets[i].tail == nil {	
		b.buckets[i].tail = n
		b.buckets[i].head = n	
		
	} else {		
		n.prev = b.buckets[i].tail
		b.buckets[i].tail.next = n
		b.buckets[i].tail = n	
	}
	b.buckets[i].mutex.Unlock()
	
}

func (b *TimedBuckets) Add(c howler.Callback, after uint64) (string, error) {
	max := b.buckets[len(b.buckets)-1].end
	if after > max  {
		return "",errors.New(fmt.Sprintf("the timeout value is greater than: %d",max))
	}
	
	n := b.newNode(&c,after)
	b.addToBuckets(n)
	// add to the hash table
	b.mutex.Lock()
	b.m[n.id] = n
	b.mutex.Unlock()	
	return n.id, nil
}

func (b *TimedBuckets) Del(id string) error {

	// get the id
	b.mutex.Lock()
	n,ok := b.m[id]
	b.mutex.Unlock()
	
	if !ok {
		return nil
	}

	// remove the node
	b.buckets[n.bucketIndex].mutex.Lock()
	
	if n.prev != nil && n.next != nil {	
		n.prev.next = n.next
		n.next.prev = n.prev
		n.next = nil
		n.prev = nil
		n = nil	
	} else if n.prev == nil {
		// we are deleting the head
		b.buckets[n.bucketIndex].head = n.next
		if n.next == nil {
			b.buckets[n.bucketIndex].tail = nil
		} else {
			n.next.prev = nil
		}	
	} else 	if n.next == nil {
		// we are deleting the tail
		b.buckets[n.bucketIndex].tail = n.prev
		n.prev.next = nil		
	}
	b.buckets[n.bucketIndex].mutex.Unlock()	
	n = nil
	return nil
}

func (b *TimedBuckets) Howls() <-chan howler.Callback {
	return b.c
}


func (b *TimedBuckets) Start() {
	// we start the timers for each bucket to
	// check for expiry (from that bucket)
	// bucket[0] is special, we need to generate event for callbacks

	// start a timer for each bucket
	for i:=1;i<len(b.buckets);i++ {
		b.buckets[i].ticker = time.NewTicker(time.Millisecond * 1)
		go b.tickHandler(i)
	}
	
}


func (b *TimedBuckets) tickHandler(bucketIndex int) {	
}

func (b *TimedBuckets) Stop() {
}
