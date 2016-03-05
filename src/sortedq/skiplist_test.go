package sortedq

import (
	"testing"
	"howler"
	"fmt"
)

func TestInsertAndCheckSort(t *testing.T) {
	insertAndCheckSort(t,200)
}


func insertAndCheckSort(t *testing.T, n int) {
	

	l := newSkiplist()

	c := new(howler.Callback)
	// insert in reverse order
	for i:=n;i>0;i=i-10 {
		l.insert(int64(i),c)
	}

	ch := make(chan int64)
	go l.yield(ch)

	prev := int64(-1000000)
	for i := range ch {		
		if prev > i {
			s := fmt.Sprintf("queue not sorted, got %d, and prev: %d",i,prev)
			t.Fatal(s)
		}
		prev = i
	}

}

func TestInsertAndCheckSortLargeSet(t *testing.T) {
	insertAndCheckSort(t,200000)
}

func TestCheckHowlList(t *testing.T) {
	
	n := 100
	l := newSkiplist()	
	for i:=n;i>0;i=i-10 {
		c := new(howler.Callback)
		c.Uri = fmt.Sprintf("uri:%d",i)
		l.insert(int64(i),c)
	}

	// insert more callbacks at 50
	for i:=0;i<9;i++ {
		c := new(howler.Callback)
		c.Uri = fmt.Sprintf("uri:%d",i)
		l.insert(int64(50),c)
	}

	// now search for 50 and check for number of callbacks
	found,prev := l.search(50)
	if !found {
		t.Fatal("did not find a node that was inserted earlier")
	}
	
	// go down
	p := prev.next
	for ;p.down!=nil; p = p.down {
	}
	
	// now loop through the callbacks
	count := 0
	for h:=p.howlHead;h!=nil;h=h.next {
		count = count + 1
	}

	if count != 10 {
		t.Fail()
	}
}

