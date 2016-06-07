package raft

type Stable interface {
	GetUint64(string) (uint64,bool)
	Get(string) (string,bool)
	Store(string,interface{})
}


