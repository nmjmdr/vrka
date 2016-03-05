package sortedq

// need to make this thread safe


import (
	"fmt"
	"math/rand"
	"howler"
)

const min = -10000000

type howlNode struct {
	c *howler.Callback
	next *howlNode
}

// the sortedq stores timestamp of the local node
// hence when a new node is being added
// it computes the timestamp by setting
// ts := after + time.Now().UnixNano()
// given this:
// the design in distributed system is this:
// 1. Every node stores the skiplist.timestamp based on its own local time
// 2. Only the master node is responsible for making the callback
// 3. The timeouts do occur on follower nodes, but they are ignored
// 4. This way we do not co-ordination across nodes

// the issue with skiplist could be: how do we scale it to millions of
// callbacks? - can we partition later?

type node struct {
	timestamp int64
	top   *node
	down *node
	next *node
	prev *node
	howlHead *howlNode
}


type list struct {
	head *node
}

func newSkiplist() *list {
	n := new(node)
	n.timestamp = min
	l := new(list)
	l.head = n
	return l
}

func (l *list) print() {

	p := l.head
	for p != nil {
		for t := p;t!=nil;t=t.next {
			fmt.Printf("%d ",t.timestamp)
		}
		fmt.Println()
		p = p.down
	}
}

func (l *list) yield(ch (chan int64)) {

	// go down
	p := l.head
	for ;p.down!=nil; p = p.down {
	}

	p = p.next
	for t := p;t!=nil;t=t.next {
		ch <- t.timestamp
	}		
	close(ch)
}


func (l *list) insert(val int64,c *howler.Callback) *node {

	found,prev := l.search(val)
	
	if !found {
		n := newNode(val,c)	
		l.insertIntoList(n,prev)
		l.promote(n)
		return n
	} else {
		addToCallbackList(prev.next,c)
		return prev.next
	}
}

func addToCallbackList(n *node,c *howler.Callback) {
	// go down to the node
	// add to the callback list there
	p := n
	for ; p.down != nil; p = p.down {		
	}

	// p.cList should never be null
	// as we would have inserted a node before
	if p.howlHead == nil {
		panic("Something went wrong - howlHead is nil")
	}

	// add at head
	hn := new(howlNode)
	hn.c = c
	hn.next = p.howlHead
	p.howlHead = hn	

}

func newNode(val int64, c *howler.Callback) *node {
	n := new(node)
	n.timestamp = val
	
	h := new(howlNode)
	h.c = c
	n.howlHead = h

	return n
}

func (l *list) insertIntoList(n *node,prev *node) {

	holdnext := prev.next
	
	prev.next = n	
	n.prev = prev
	n.next = holdnext
	if holdnext != nil {
		holdnext.prev = n
	}	
}

func (l *list) promote(n *node) {

	for {
		if toss() == false {		
			return
		}

		//fmt.Printf("promoting: %d\n",n.timestamp)
		
		prevAtNextLevel := l.getPrevAtNextLevel(n)

		m := new(node)
		m.down = n
		m.timestamp = n.timestamp
		n.top = m

		holdnext := prevAtNextLevel.next
		prevAtNextLevel.next = m
		m.prev = prevAtNextLevel
		m.next = holdnext
		if holdnext != nil {
			holdnext.prev = m
		}
		
		n = m
	}
}


func (l *list) getPrevAtNextLevel(n *node) *node {

	p := n

	for p != nil {
		if p.top != nil {
			return p.top
		}
		p = p.prev
	}

	// we reached the end and could not find a previous with a top set
	// create one at head 
	h := new(node)
	h.timestamp = l.head.timestamp
	h.down = l.head
	l.head.top = h
	l.head = h

	return h
}

func toss() bool {
	v := rand.Intn(2)
	if v == 1 {
		return true
	}
	return false
}

// returns the node previous to node that was found or not found
func (l *list) search(val int64) (bool,*node) {

	levelp := l.head

	for {		
		found,prev := trace(val,levelp)
		if found {
			return true,prev
		} 

		if prev.down == nil {
			// we have reached the last level
			// return
		return false, prev
		}	
		levelp = prev.down
	}
}

func  trace(val int64,t *node) (bool,*node) {
		
	prev := t
	for t !=nil && t.timestamp <= val {

		if t.timestamp == val {
			return true,prev
		}
		prev = t
		t = t.next
	}
	//fmt.Printf("Trace: returning: %d\n",prev.timestamp)
	return false,prev
}

