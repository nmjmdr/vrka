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
	c  chan howler.Callback
	buckets []bucket
	m map[string](*node)
	flk *flake.Flk
	mutex *sync.Mutex
	createTw tickerCreator

	// bucket levels and other factors
	numBuckets int //default: 8
	factor int //default 16
	firstBucketStart int //default 0
	firstBucketEnd int //default : 256
	sweepDuration int // default: 1 in milliseonds	
}
func (b *TimedBuckets) setDefaults() {
	// bucket levels and other factors
	b.numBuckets = 8       //default: 8
	b.factor =16           //default 16
	b.firstBucketStart = 0 //default 0
	b.firstBucketEnd = 256 //default : 256
	b.sweepDuration = 1    // default: 1 in milliseonds
}

func (b *TimedBuckets) setup() {
	b.buckets = make([]bucket,b.numBuckets)

	start := uint64(b.firstBucketStart)
	end := uint64(b.firstBucketEnd)

	b.buckets[0].start = start
	b.buckets[0].end = end
	b.buckets[0].mutex = &sync.Mutex{}
	b.buckets[0].tw = b.createTw(time.Duration(b.sweepDuration)*time.Millisecond)

	for n:=1;n<b.numBuckets;n++ {
		start = end+1
		end = end * uint64(b.factor)
		b.buckets[n].start = start
		b.buckets[n].end = end	
		b.buckets[n].mutex = &sync.Mutex{}
		b.buckets[n].tw = b.createTw(time.Duration(b.sweepDuration)*time.Millisecond)		
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
	b.setDefaults()
	ctor(createTw,b)
	return b
}


func NewBucketsCustomLevels(createTw tickerCreator,numBuckets int, factor int) *TimedBuckets {
	b := new(TimedBuckets)
	b.setDefaults()
	b.factor = factor
	b.numBuckets = numBuckets
	ctor(createTw,b)
	return b
}

func ctor(createTw tickerCreator, b *TimedBuckets) {
	b.createTw = createTw
	b.setup()
	var err error
	b.flk,err = flake.FlakeNode()
	b.m = make(map[string](*node))
	b.mutex = &sync.Mutex{}	
	if err != nil {
		panic(err)
	}
	b.c = make(chan howler.Callback)
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

func (b *TimedBuckets) disconnectNode(n *node) *node  {

	holdNext := n.next

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
	return holdNext
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
			b.checkForExpiry()
		} else {
			// sweep the bucket normally
			b.sweepBucket(bucketIndex)
		}
		
        }
}
func (b *TimedBuckets) checkForExpiry() {
	// lock the bucket for sweeping
	
	b.buckets[0].mutex.Lock()
	
	for p := b.buckets[0].head; p != nil ; {
		// reduce the time		
		p.after = (p.after - uint64(b.sweepDuration))
		// check for expiry
		if p.after <= 0 {
			// publish
			b.c <- (*p.c)	
			// delete and set the next one to check
			p  = b.disconnectNode(p)
		} else {
			p = p.next
		}
	}
	b.buckets[0].mutex.Unlock()
}


func (b *TimedBuckets) sweepBucket(bIndex int) {
	
	// lock the bucket for sweeping
	b.buckets[bIndex].mutex.Lock()
	
	for p := b.buckets[bIndex].head; p!=nil; {	
		// reduce the time by sweep duration
		p.after = (p.after - uint64(b.sweepDuration))
		// check if it has to moved
		if p.after < b.buckets[bIndex].start {
			// move up and set the next one to sweep
			p = b.moveUp(p)
		} else {
			p = p.next
		}
	}
	b.buckets[bIndex].mutex.Unlock()
}

func (b *TimedBuckets) moveUp(n *node) *node {
	
	next := b.disconnectNode(n)
	// move it up one level
	n.bucketIndex = n.bucketIndex - 1	
	// add the node to new level
	b.addToBucket(n)
	return next
}

func (b *TimedBuckets) Stop() {
	// stop the timer for each bucket
	for i:=1;i<len(b.buckets);i++ {
		b.buckets[i].tw.Stop()		
	}
}
