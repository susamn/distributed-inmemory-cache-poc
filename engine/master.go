package engine

import (
	"distributed-inmemory-cache/config"
	"distributed-inmemory-cache/model"
	"fmt"
	"log"
	"strconv"
	"time"
)

type Master struct {
	data          map[string]string
	dataVersionId int64
	nodes         []*Slave
	MasterPort    int
	nextNodePort  int
}

func NewMaster(config *config.Config) *Master {
	master := &Master{
		data:          make(map[string]string),
		dataVersionId: time.Now().UnixMilli(),
		MasterPort:    config.Service.Master.Port,
		nextNodePort:  config.Service.Master.NodePortInitial,
		nodes:         make([]*Slave, 0, config.Service.Nodes.MinCount),
	}

	master.tryRecoveringNodes()

	if len(master.nodes) <= config.Service.Nodes.MinCount {
		fmt.Println("Scaling to meet minimum node count")
		for i := 0; i <= config.Service.Nodes.MinCount-len(master.nodes); i++ {
			master.ScaleUp(config)
		}
	}

	return master

}

func (master *Master) tryRecoveringNodes() {
	fmt.Println("Trying to recover nodes")
	for i := 0; i < 20; i++ {
		port := master.nextNodePort
		existingNode := ExistingNode(port, master)
		if existingNode.Status == Active {
			master.nodes = append(master.nodes, existingNode)
			master.nextNodePort = master.nextNodePort + 1
			fmt.Println("Recovered 1 node with port: " + strconv.Itoa(port))
		}
	}
	var nodeData map[string]string
	var currentVersion int64
	if len(master.nodes) > 0 {
		for _, node := range master.nodes {
			if node.DataVersionId > currentVersion {
				d, err := node.GetData()
				if err == nil {
					nodeData = d.Data
				}
			}
		}
	}
	if nodeData != nil {
		master.data = nodeData
	}
}

func (master *Master) AddNode(config *config.Config) error {
	if len(master.nodes) < config.Service.Nodes.MaxCount {
		node := NewNode(master.nextNodePort, master)
		master.nodes = append(master.nodes, node)
		master.nextNodePort = master.nextNodePort + 1
	}
	return nil
}

func (master *Master) Broadcast() {
	log.Println("Master: Sending broadcast")
	for _, node := range master.nodes {
		err := node.Broadcast(master.dataVersionId)
		if err != nil {
			fmt.Println("Could not broadcast to node: " + strconv.Itoa(node.Port))
			fmt.Println(err)
		}
	}
}

func (master *Master) GetData() map[string]string {
	return master.data
}

func (master *Master) GetReplicationData() *model.DataPayload {
	return &model.DataPayload{DataVersion: master.dataVersionId, Data: master.data}
}

func (master *Master) SetData(data map[string]string) {
	for k, v := range data {
		master.data[k] = v
	}
	master.dataVersionId = time.Now().UnixMilli()
	master.Broadcast()
}

func (master *Master) DeleteData(data []string) {
	for _, val := range data {
		delete(master.data, val)
	}
	master.dataVersionId = time.Now().UnixMilli()
	master.Broadcast()
}

func (master *Master) ScaleUp(conf *config.Config) bool {
	if len(master.nodes) < conf.Service.Nodes.MaxCount {
		node := NewNode(master.nextNodePort, master)
		node.Start()
		master.nodes = append(master.nodes, node)
		master.nextNodePort = master.nextNodePort + 1
		<-time.After(3 * time.Second)
		master.Broadcast()
		master.refreshNodes()
		return true
	}
	return false
}

func (master *Master) ScaleDown(conf *config.Config) bool {
	if len(master.nodes) > conf.Service.Nodes.MinCount {
		node := master.nodes[0]
		err := node.Shutdown()
		if err != nil {
			fmt.Println("Could not shutdown node: " + strconv.Itoa(node.Port) + "Error")
			return false
		}
		master.nodes = master.nodes[1:]
		master.refreshNodes()
		return true
	}
	return false
}

func (master *Master) refreshNodes() {
	log.Println("Master: Node refresh running")
	for _, node := range master.nodes {
		node.Refresh(master.dataVersionId)
	}
}

func (master *Master) NodeStats() map[string]interface{} {
	response := make(map[string]interface{})
	response["nodeCount"] = len(master.nodes)
	response["dataVersionId"] = master.dataVersionId
	response["nodes"] = master.nodes
	return response
}

func (master *Master) MakeAvailable() {
	log.Println("Master: made available")
	<-time.After(2 * time.Second)
	for _, node := range master.nodes {
		if node.Status == New {
			node.Start()
		} else {
			log.Println("Node " + strconv.Itoa(node.Port) + " is still running, will recover it")
			node.Status = Recovered
		}
	}
	<-time.After(3 * time.Second)
	master.Broadcast()
	master.refreshNodes()
}

func (master *Master) KillAllNodes() error {
	log.Println("Master: kill all nodes triggered")
	for _, node := range master.nodes {
		err := node.Shutdown()
		if err != nil {
			log.Fatal(fmt.Sprintf("Node with port %d could not be stopped", node.Port))
		}
	}
	return nil
}
