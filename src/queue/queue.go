package queue

import (
	"common"
	"sync/atomic"
	"time"
	"math/rand"	
)

const defaultQSize = 1000
const maxUint64 = 18446744073709551615 

type Queue interface {
	Add(at uint64,reminder *common.Reminder)
	GetMin() *common.Reminder
	GetMinInRange(min uint64,max uint64) (*common.Reminder,bool)
}


type multiQueue struct {
	qs []PriorityQueue
	locks []int32
	n int
	rd *rand.Rand
	numItems int32
}

func NewQueue(n int) Queue {
	m := new(multiQueue)
	m.setup(n)

	seed := rand.NewSource(time.Now().UnixNano())
	m.rd = rand.New(seed)
	
	return m
}

func (m *multiQueue) setup(n int) {
	m.n = n
	m.qs = make([]PriorityQueue,n)
	m.locks = make([]int32,n)
	for i:=0;i<n;i++ {
		m.qs[i] = NewPQ(defaultQSize)
	}
}

func (m *multiQueue) Add(at uint64,reminder *common.Reminder) {

	for {
		i := m.getRandom()
		// try lock i
		if atomic.CompareAndSwapInt32(&(m.locks[i]),0,1) {
			// locked
			item := item{}
			item.at = at
			item.r = reminder
			m.qs[i].Add(item)
			// unlock
			atomic.StoreInt32(&(m.locks[i]),0)
			// increment the number of items
			atomic.AddInt32(&(m.numItems),1)
			break
		}
	}	
}

func (m *multiQueue) getRandom() int {
	return m.rd.Intn(m.n)
}

// Here is how this function works:
// 1. It reads the heads of all queues
// 2. Checks if any of them are in the give range
// 3. If any of the values are within range, it
// attempts to lock it and delete it, if it cannot
// it attemps to lock the next one in the range (and so on)
// if it is unable to obtain lock on any
// (or alternatively no value is in min) the method returns false
// once the method sucessfully obtains a lock
// it deletes the element and returns
func (m *multiQueue) GetMinInRange(min uint64,max uint64) (*common.Reminder,bool) {

	
	for i:=0;i<m.n;i++ {
		v,flag := m.qs[i].GetHead()
		if !flag {
			continue
		}
		if (min <= v) && (v<=max) {
			it,flag := m.TryAndDelMin(i)
			if !flag {
				continue
			} else {
				return it.r,true
			}
		}
	}
	return nil,false
}

func (m *multiQueue) GetMin() *common.Reminder {

	
	var flagI,flagJ bool
	var iMin,jMin uint64
	
	for {
		if atomic.LoadInt32(&(m.numItems)) == 0 {
			return nil
		}
		
		i := m.getRandom()
		j := m.getRandom()

		iMin,flagI = m.qs[i].GetHead()
		jMin,flagJ = m.qs[j].GetHead()

		if flagI == false && flagJ == false {
			continue
		}

		if flagI == false && flagJ == true {
			i = j
		} else {
			if iMin > jMin {
				i = j
			}
		}
		
		
		// try lock i
		it,flag := m.TryAndDelMin(i)
		if !flag {
			continue
		} else {
			return it.r
		}
	}
	
	
}

func (m *multiQueue) TryAndDelMin(i int) (item,bool)  {

	if atomic.CompareAndSwapInt32(&(m.locks[i]),0,1) {
		// locked
		it,flag := m.qs[i].DelMin()
		if flag == false {
			return item{},false
		}
		// reduce the number of items
		atomic.AddInt32(&(m.numItems),-1)
		// unlock
		atomic.StoreInt32(&(m.locks[i]),0)
		return it,true
	} else {
		return item{},false
	}
}
