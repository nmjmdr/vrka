package raft

import (
	"fmt"
)

type inMemoryStable struct {
	m map[string]interface{}
}


func newInMemoryStable() *inMemoryStable {
	i :=new(inMemoryStable)
	i.m = make(map[string]interface{})

	return i
}

func (i *inMemoryStable) GetUint64(key string) (uint64,bool) {

	v,ok := i.m[key]

		
	if !ok {
		return uint64(0),false
	}

	var u uint64
	u,ok = v.(uint64)
		
	if !ok {
		return uint64(0),false
	}

	return u,true
}


func (i *inMemoryStable) Get(key string) (string,bool) {

	v,ok := i.m[key]

	if !ok {
		return "",false
	}

	var u string
	u,ok = v.(string)

	if !ok {
		return "",false
	}

	return u,true
}


func (i *inMemoryStable) Store(key string,value interface{}) {
	i.m[key] = value
	fmt.Println(i.m)
}
