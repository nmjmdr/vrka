package raft

type Stable interface {
	GetUint64(string) (uint64,error)
}
