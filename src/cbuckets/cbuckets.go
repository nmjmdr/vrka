package cbuckets

import (
	"howler"	
	"sync"
	"flake"
	"errors"
	"fmt"
	"tickerwrap"
	"time"
)

const numBuckets = 8
const factor = 16
const firstBucketStart = 0
const firstBucketEnd = 256
const sweepDuration = 1 // in milliseonds

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
	tw tickerwrap.Tickerw
}

type tickerCreator func(d time.Duration) tickerwrap.Tickerw


type TimedBuckets struct {
	c  (<-chan howler.Callback)
	buckets []bucket
	m map[string](*node)
	flk *flake.Flk
	mutex *sync.Mutex
	createTw tickerCreator
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
		b.buckets[n].tw = b.createTw(sweepDuration*time.Millisecond)		
	}
}


func (b *TimedBuckets) newId() string {
	id,err := b.flk.NextId()
	if err != nil {
		panic(err)
	}
	return string(id)
}

func NewBuckets(createTw tickerCreator) *TimedBuckets {
	b := new(TimedBuckets)
	b.createTw = createTw
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
	b.buckets[n.bucketIndex].mutex.Lock()
	b.addToBucket(n)
	b.buckets[n.bucketIndex].mutex.Unlock()	
}

func (b *TimedBuckets) addToBucket(n *node) {
	

	if b.buckets[n.bucketIndex].tail == nil {	
		b.buckets[n.bucketIndex].tail = n
		b.buckets[n.bucketIndex].head = n	
		
	} else {		
		n.prev = b.buckets[n.bucketIndex].tail
		b.buckets[n.bucketIndex].tail.next = n
		b.buckets[n.bucketIndex].tail = n	
	}

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

func (b *TimedBuckets) disconnectNode(n *node) {

	if n.prev != nil && n.next != nil {	
		n.prev.next = n.next
		n.next.prev = n.prev
		n.next = nil
		n.prev = nil			
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
	b.disconnectNode(n)
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
	for i:=0;i<len(b.buckets);i++ {
		b.buckets[i].tw.Start()
		go b.tickHandler(i)
	}	
}


func (b *TimedBuckets) tickHandler(bucketIndex int) {

	// need to wait on the ticker channel
	for range b.buckets[bucketIndex].tw.Channel() {
		if bucketIndex == 0 {
		} else {
			// sweep the bucket normally
			b.sweepBucket(bucketIndex)
		}
        }
}


func (b *TimedBuckets) sweepBucket(bIndex int) {
	
	// lock the bucket for sweeping
	b.buckets[bIndex].mutex.Lock()
	for p := b.buckets[bIndex].head; p!=nil;p=p.next {	
		// reduce the time by sweep duration
		p.after = (p.after - sweepDuration)
		// check if it has to moved
		if p.after < b.buckets[bIndex].start {
			b.moveUp(p)
		}
	}
	b.buckets[bIndex].mutex.Unlock()
}

func (b *TimedBuckets) moveUp(n *node) {
	
	b.disconnectNode(n)
	// move it up one level
	n.bucketIndex = n.bucketIndex - 1
	// add the node to new level
	b.addToBucket(n)
}

func (b *TimedBuckets) Stop() {
	// stop the timer for each bucket
	for i:=1;i<len(b.buckets);i++ {
		b.buckets[i].tw.Stop()		
	}
}
