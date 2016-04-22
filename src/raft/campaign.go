package raft

import (
	"sync"
)


type Campaign interface {
	Start(currentTerm uint32) 
	Stop()
	Elected() chan bool
}

type campaign struct {
	votesGot uint32
	config *Config
	nodeId string
	elected chan bool
	mutex *sync.Mutex
}

func NewCampaign(config *Config,nodeId string) Campaign  {
	c := new(campaign)
	c.mutex = &sync.Mutex{}
	c.votesGot = 0
	c.nodeId = nodeId
	c.config = config
	return c
}

func (c *campaign) Elected() chan bool {
	return c.elected
}

func (c *campaign) Start(currentTerm uint32) {
}

func (c *campaign) Stop() {
}
