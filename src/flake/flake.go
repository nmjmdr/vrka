package flake

import (
	"time"
	"sync"
	"errors"
	"os"
	"math"
	"macNodeId"
	"encoding/binary"
)

//
// Acknowledgement: Based on Factual\Skuld's implementation of Flake Id
// 160 bit flake id

/*
  [64 bits | Timestamp, in milliseconds since the epoch]
  [32 bits | a per-process counter]
  [48 bits | a host identifier]
  [16 bits | the process ID]
*/

/* follows a big endian approach to the id */
/* this means:
  64 --> byte[0] to byte[7](inclusive)
  32 --> byte[8] to byte [11](inclusive)
  48 --> byte[12] byte[17](inclusive)
  16 --> byte[18] to byte[19](inclusive)
*/

// Consider the time to start from year 2010
// time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)
const epoch int64 = 1262304000000000000
const maxCounter = math.MaxUint32

// node id length is 6 bytes
const nodeIdLength = 6


var nodeId []byte
var initialized bool
var initMutex = &sync.Mutex{}


/* 
 Process Id related assumptions:
 the assumption here is that process ids will not execced int16
 that is max: 65535
 but they can, check: http://blogs.msdn.com/b/oldnewthing/archive/2014/02/05/10495426.aspx

 code might generate overallping process ids, if process ids get larger than: 65535
 also it is assumed that process ids cannot be negative
*/


type Flk struct {
	mutex *sync.Mutex
	lastNow int64
	counter uint32
}

type NodeId  interface {
	// a 48 byte node id
	Id() ([]byte,error)
}

func newFlk() *Flk {
	f :=new(Flk)
	f.mutex = &sync.Mutex{}
	return f
}


func FlakeNodeCustom(nId NodeId) (*Flk,error) {
	initMutex.Lock()
	defer initMutex.Unlock()
	var err error	
	if !initialized {
		nodeId,err = nId.Id()
		initialized = true
	}
	return newFlk(),err
}

func FlakeNode() (*Flk,error) {
	macId := new(macNodeId.MacNodeId) 
	f,err := FlakeNodeCustom(macId)
	
	return f,err
}



func (f *Flk) NextId() ([]byte,error) {

	b := make([]byte,20)
	
	//copy the timestamp
	now := time.Now().UnixNano()
	if now < f.lastNow {
		// time has moved backwards, this could lead to issues, error out
		return nil,errors.New("Time has moved back, cannot proceed further")
	}

	lastTs := (f.lastNow - epoch)/1000000
	ts := (now - epoch)/1000000

	f.mutex.Lock()
	// is this really neccsarry?
	// we are ending up probably locking the whole block more than necessary
	// this is an edge case: in case there have been equal or more calls
	// than 2^32, then we need to wait until we move to next timestamp
	// is the counter nearing a reset to zero?
	// and if our lastTs the same as current ts?
	if (f.counter == maxCounter) && (ts <= lastTs) {
		// then we need wait until ts is greater than
		// lastTs - so that we do not end up repeatingids for the current node
		for ;ts <= lastTs; {
		now = time.Now().UnixNano()
		ts = (now - epoch)/1000000
		} 
	}
	 
	binary.BigEndian.PutUint64(b[0:8],uint64(ts))
	
	//increment the counter and assign to bytes
	// has time moved since the last call?
	if lastTs < ts {
		// reset the counter
		f.counter = 0
	} // if the time has not moved since the last call, we just increment
	// the counter
	
	f.counter++
	binary.BigEndian.PutUint32(b[8:12],f.counter)

	f.lastNow = now    

	f.mutex.Unlock()

	//copy the host id or node id
	copy(b[12:(12+nodeIdLength)],nodeId)
	
	//copy the process id
	pid := uint16(os.Getpid())	
	binary.BigEndian.PutUint16(b[18:20],pid)
	return b,nil
}



