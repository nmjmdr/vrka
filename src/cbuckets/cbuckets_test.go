package cbuckets

import (
	"testing"
	"howler"
	"sync"
	"tickerwrap"
	"time"
)

func createTw(d time.Duration) tickerwrap.Tickerw   {
	return tickerwrap.NewMockTicker()
}


func TestBucketSetup(t *testing.T) {
	cb := NewBuckets(createTw)

	for i:=1;i<len(cb.buckets);i++ {
		end := (cb.buckets[i].start-1)*factor
		if end != cb.buckets[i].end {
			t.Fail()
		}
	}
}


func TestTailSet(t *testing.T) {
	
	b := NewBuckets(createTw)
	cb := howler.Callback{}
	cb.Uri = "uri"
	cb.Payload = "payload"
	b.Add(cb,1)
	
	if b.buckets[0].tail == nil {
		t.Fail()
	}
}


func TestTailSetParallel(t *testing.T) {
	
	b := NewBuckets(createTw)
	cb := howler.Callback{}
	cb.Uri = "uri"
	cb.Payload = "payload"

	done := make(chan bool)
	running := make(chan bool)
	n := 10
	
	var m sync.Mutex
 	cond := sync.NewCond(&m)
	allset := false

	for i:=0;i<n;i++ {
		go func() {				
			cond.L.Lock()
			running <- true
			for !allset {
				cond.Wait()
			}
			cond.L.Unlock()
			b.Add(cb,uint64(i))
			done <- true
			
		}()			
	}	

	// wait for everything to start running
	for i:=0;i<n;i++ {
		<-running
		if i == n-1 {
			cond.L.Lock()
			allset = true
			cond.L.Unlock()
		}
	}
	
	// now signal all go routines to proceed with their work
	cond.L.Lock()
	cond.Broadcast()
	cond.L.Unlock()
	
	// wait for everyting to finish
	for i:=0;i<n;i++ {
		<-done		
	}
	
	if b.buckets[0].tail == nil {
		t.Fail()
	}

	// count the number of entries, it should be equal to n
	i := 0
	for p:=b.buckets[i].head;p!=nil;p=p.next {
		i++
	}

	if i != n {
		t.Fail()
	}
}

func createWithAllBucketsSet(t *testing.T) (*TimedBuckets,[]string) {
	b := NewBuckets(createTw)
	cb := howler.Callback{}
	cb.Uri = "uri"
	cb.Payload = "payload"

	ids := make([]string,len(b.buckets))

	var err error
	for i := 0;i<len(b.buckets);i++ {
		ids[i],err =  b.Add(cb,(b.buckets[i].start + 1))
		if err != nil {
			t.Fatal(err)
		}
	}

	return b,ids[:]
}

func TestAllBucketsInsert(t *testing.T) {

	b,_ := createWithAllBucketsSet(t)

	for i := 0;i<len(b.buckets);i++ {
		if b.buckets[i].head == nil {
			t.Fail()
		}
	}
}


func TestDelete(t *testing.T) {

	b,ids := createWithAllBucketsSet(t)
	
	for i:=0;i<len(ids);i++ {
		b.Del(ids[i])
	}

	for i:=0;i<len(b.buckets);i++ {
		if b.buckets[i].head != nil {
			t.Fail()
		}
	}

}

func TestAddDelParallel(t *testing.T) {

	b := NewBuckets(createTw)
	perBucket := 1000
	ids := make([]string,len(b.buckets)*perBucket)

	cb := howler.Callback{}
	cb.Uri = "uri"
	cb.Payload = "payload"
	index := 0
	var err error
	for i:=0;i<len(b.buckets);i++ {
		for j:=0;j<perBucket;j++ {
			ids[index],err = b.Add(cb,b.buckets[i].start + uint64(1 + j) )
			if err != nil {
				t.Fatal(err)
			}
			index = index + 1
		}
	}

	t.Log(count(b))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for i:=0;i<len(ids);i++ {
			b.Del(ids[i])
		}
		wg.Done()
	}()

	
	go func() {
		// start adding in parallel
		for i:=0;i<len(b.buckets);i++ {
			for j:=0;j<perBucket;j++ {
				b.Add(cb,b.buckets[i].start + uint64(1 + j))			
			}
		}
		wg.Done()
	}()


	wg.Wait()
	

	// now there should only be perBucket * num-buckets items
	if count(b) != len(b.buckets)*perBucket {
		t.Fail()
	}
}

func count(b *TimedBuckets) int {

	count := 0
	for i:=0;i<len(b.buckets);i++ {
		for p := b.buckets[i].head;p!=nil;p=p.next {
			count++
		}
	}
	return count
}

