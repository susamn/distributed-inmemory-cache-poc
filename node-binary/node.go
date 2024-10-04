package main

import "time"

type DataPayload struct {
	DataVersion  int64             `json:"data_version"`
	Data         map[string]string `json:"data"`
	PID          int               `json:"pid"`
	RunningSince int64             `json:"running_since"`
}

type Node struct {
	Data            map[string]string
	DataVersion     int64
	NodePort        int
	MasterPort      int
	ShutdownChannel chan bool
	PID             int
	RunningSince    int64
}

func NewNode(nodePort int, masterPort int, shutdownChannel chan bool, pid int) *Node {
	return &Node{
		Data:            make(map[string]string),
		NodePort:        nodePort,
		MasterPort:      masterPort,
		ShutdownChannel: shutdownChannel,
		PID:             pid,
		RunningSince:    time.Now().UnixMilli(),
	}
}
