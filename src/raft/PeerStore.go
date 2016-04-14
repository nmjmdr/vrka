package raft

import (	
	"os"	
	"encoding/json"
)

type PeerStore interface {
	Get() ([]Peer,error)
}

type peerStore struct {
	path string
}

func NewPeerStore(path string) PeerStore {
	p := new(peerStore)
	p.path = path
	return p
}

func (p *peerStore) Get() ([]Peer,error) {
	return p.loadPeers()
}

func (p *peerStore) loadPeers() ([]Peer,error) {
	// read from file
	peers := make([]Peer,0)
	file,err := os.Open(p.path)
	if err != nil {
		return nil,err
	}
	
	err = json.NewDecoder(file).Decode(&peers)
	if err != nil {
		return nil,err
	}
	
	return peers,err
}

