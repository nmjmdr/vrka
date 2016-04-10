package raft

import (
	"io"
	"encoding/json"
	"os"
	"bufio"	
)

type PeerStore interface {
	Get() []Peer
	Add(p Peer) []Peer
	Set(peers []Peer)
	Changed() <-chan bool
}

type peerStore struct {
	changed chan bool
	reader io.Reader
	writer io.Writer
}

func NewPeerStore(path string) (PeerStore,error) {

	file,err := os.Open(path)
	if err != nil {
		return nil,err
	}

	r := bufio.NewReader(file)
	w := bufio.NewWriter(file)
	f := NewPeerStoreRW(r,w)
	return f,nil
}

func NewPeerStoreRW(r io.Reader,w io.Writer) PeerStore {
	f := new(peerStore)
	f.changed = make(chan bool)
	f.reader = r
	f.writer = w	
	return f
}

func (f *peerStore) Changed() <-chan bool {
	return f.changed
}

func (f *peerStore) Get() []Peer {
	var peers []Peer	
	json.NewDecoder(f.reader).Decode(&peers)	
	return peers
}

func (f *peerStore) Add(p Peer) []Peer {
	return nil
}


func (f *peerStore) Set(peers []Peer) {
	if peers == nil {
		return
	}	
	json.NewEncoder(f.writer).Encode(peers)	
}
