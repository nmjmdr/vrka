package queue

import (
	"testing"
	"common"
	"strconv"
	"sync"
)


func Test_Add(t *testing.T) {
	q := NewPQ(5)
	for i:=5;i>=1;i-- {
		t:= item{}
		t.at = uint64(i)
		q.Add(t)
	}

	v,e := q.GetHead()
	if e != true {
		t.Fatal("queue is empty")
	}	
	if v != 1 {
		t.Fail()
	}
}


func Test_QFull(t *testing.T) {
	pq := NewPQ(5)
	q,_ := pq.(*dArr)
	for i:=5;i>=1;i-- {
		t:= item{}
		t.at = uint64(i)
		q.Add(t)
	}

	if !q.isFull() { 
		t.Fail()
	}
}

func Test_Resize(t *testing.T) {
	pq := NewPQ(5)
	q,_ := pq.(*dArr)
	for i:=10;i>5;i-- {
		t:= item{}
		t.at = uint64(i)
		q.Add(t)
	}

	// the queue should be full at this point
	if !q.isFull() {
		t.Fatal("Expecting queue to be full, but its not")
	}

	it := item{}
	it.at = 1
	q.Add(it)
	min,_ := q.GetHead()
	if min != 1 {
		t.Fail()
	}
		
	
}

func Test_DelMin(t *testing.T) {
	q := NewPQ(10)
	for i:=10;i>=1;i-- {
		t:= item{}
		t.at = uint64(i)
		q.Add(t)
	}
	
	for i:=1;i<=10;i++ {
		r,e := q.DelMin()
		if e != true {
			t.Fatal("queue is empty")
		}
		t.Log(r.at)
		if uint64(i) != r.at {
			t.Fatal("unexpected output from queue")
		}
	}
}

func Test_SimpleMQ(t *testing.T) {
	addRemoveMultiQueue(t,2,10)
}

func Test_10QMQ(t *testing.T) {
	addRemoveMultiQueue(t,10,1000)
}

func addRemoveMultiQueue(t *testing.T,nQueues int,max int) {

	q := NewQueue(nQueues)

	for i:=0;i<max;i++ {
		r := new(common.Reminder)
		r.Payload = strconv.Itoa(i)
		q.Add(uint64(i),r)
	}

	for i:=0;i<max;i++ {
		d := q.GetMin()
		t.Log(d)
	}	
}


func Test_ParallelMQ(t *testing.T) {

	nQueues := 5
	q := NewQueue(nQueues)

	// start nQueue rountines that insert n items parallely
	n := 1000
	step := n/nQueues

	var wg sync.WaitGroup
	wg.Add(nQueues)
	for p:=0;p<nQueues;p++ {
		go func(part int) {			
			start := p * step
			end := (p+1) * step
			for i:=start;i<end;i++ {
				t.Log(i)
				r := new(common.Reminder)
				r.Payload = strconv.Itoa(i)
				q.Add(uint64(i),r)
			}
			wg.Done()
		}(p)
	}
	wg.Wait();

	wg.Add(nQueues)
	for p:=0;p<nQueues;p++ {
		go func(part int) {			
			start := p * step
			end := (p+1) * step
			for i:=start;i<end;i++ {
				v := q.GetMin()
				t.Log(v)
			}
			wg.Done()
		}(p)
	}
	wg.Wait();
	
}

func Test_DelInRange(t *testing.T) {

	nQueues := 5
	q := NewQueue(nQueues)

	max := 100
	for i:=0;i<max;i++ {
		r := new(common.Reminder)
		r.Payload = strconv.Itoa(i)
		q.Add(uint64(i),r)
	}

	count :=0
	for i:=0;i<max-1;i++ {
		for j:=0;j<=i+1;j++ {
			d,f := q.GetMinInRange(0,uint64(j))
			if f {
				t.Log(d.Payload)
				count = count + 1
			}
		}
	}
	if count != max {
		t.Fail()
	}
}

