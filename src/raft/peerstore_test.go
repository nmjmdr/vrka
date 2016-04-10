package raft

import (
	"testing"
	"bufio"	
	"bytes"	
)


func Test_SetGet(t *testing.T) {

	p := make([]Peer,2)	
	p[0].Id = "1"
	p[1].Id = "2"

	buf := bytes.NewBuffer(nil)
	writer := bufio.NewWriter(buf)
	reader := bufio.NewReader(buf)

	peerstore := NewPeerStoreRW(reader,writer)
	peerstore.Set(p)
	writer.Flush()
	
	actual := peerstore.Get()

	if actual == nil || len(actual) != 2 {
		t.Fatal("Unable to set - actaul is nil or less than expected length")
	}

	for i:=0;i<len(actual);i++ {
		if actual[i].Id != p[i].Id {
			t.Fatal("Unable to set - value is different")
		}
	}	
}


