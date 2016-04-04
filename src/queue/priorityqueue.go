package queue

import (
	"common"
)

type item struct {
	at uint64
	r *common.Reminder
}

type PriorityQueue interface {
	Add(i item)
	GetHead() (uint64,bool)
	DelMin() (item,bool)
}


type dArr struct {
	arr []item
	n int
}

func NewPQ(sz int) PriorityQueue {
	d := new(dArr)
	// array starts from 1
	d.arr = make([]item,(sz+1))	
	// to begin with 0 items
	d.n = 0
	return d
}


func (d *dArr) Add(i item)  {
	// insert at the end, and then swim up
	if d.isFull() {
		d.resize()
	}

	d.n = d.n+1
	d.arr[d.n] = i	
	d.swim(d.n)
	
}

func (d *dArr) isFull() bool {
	return d.n == (len(d.arr)-1)
}

func (d *dArr) resize() {
	// allocate new array
	r := make([]item,len(d.arr)*2)
	// copy all the elements
	for i:=0;i<len(d.arr);i++ {
		r[i] = d.arr[i]
	}
	d.arr = r
}

func (d *dArr) swim(i int) {

	for i > 1 && d.arr[i].at < d.arr[i/2].at {
		d.swap(i,i/2)
		i = i/2
	}
}

func (d *dArr) swap(i int, j int) {
	hold := d.arr[i]
	d.arr[i] = d.arr[j]
	d.arr[j] = hold
}

func (d *dArr) GetHead() (uint64,bool) {
	if d.n == 0 {
		return 0,false
	}
	return d.arr[1].at,true
}

func (d *dArr) DelMin() (item,bool) {
	if d.n == 0 {
		return item{},false
	}	
	i := d.arr[1]
	// put the last one at head and then sink it
	d.arr[1] = d.arr[d.n] 
	d.n = d.n - 1
	d.sink(1)
	return i,true
}


func (d *dArr) sink(k int) {
	for 2*k <= d.n {	
		left := 2*k
		right := (2*k + 1)

		j := right
		if right > d.n || d.arr[left].at < d.arr[right].at {
			j = left
		}		
		if d.arr[k].at < d.arr[j].at {
			break
		}

		d.swap(k,j)
		k = j
	}
}
