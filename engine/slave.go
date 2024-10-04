package engine

import (
	"distributed-inmemory-cache/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

type NodeStatus int

const (
	New NodeStatus = iota
	Active
	Shutdown
	Zombie // 6
	Recovered
	Unrecoverable
)

type NodeDataQuality int

const (
	Fresh NodeDataQuality = iota
	Dirty
)

type Slave struct {
	DataVersionId  int64 `json:"dataVersionId"`
	Port           int   `json:"port"`
	master         *Master
	binary         string
	Status         NodeStatus `json:"status"`
	broadcastURL   string
	killURL        string
	healthURL      string
	dataVersionURL string
	dataUrl        string
	ProcessId      int             `json:"processId"`
	RunningSince   int64           `json:"runningSince"`
	DataQuality    NodeDataQuality `json:"dataQuality"`
}

func NewNode(port int, master *Master) *Slave {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return nil
	}

	binaryPath := filepath.Join(wd, "node-binary", "node")
	broadcastURL := fmt.Sprintf("http://localhost:%d/notify", port)
	killURL := fmt.Sprintf("http://localhost:%d/kill", port)
	healthURL := fmt.Sprintf("http://localhost:%d/health", port)
	dataUrl := fmt.Sprintf("http://localhost:%d/data", port)
	dataVersionURL := fmt.Sprintf("http://localhost:%d/dataVersion", port)
	node := &Slave{
		Port:           port,
		master:         master,
		binary:         binaryPath,
		broadcastURL:   broadcastURL,
		killURL:        killURL,
		healthURL:      healthURL,
		dataUrl:        dataUrl,
		dataVersionURL: dataVersionURL,
		DataQuality:    Dirty,
		Status:         New,
	}
	return node
}

func ExistingNode(port int, master *Master) *Slave {
	node := NewNode(port, master)
	node.Status = Zombie
	if node.CheckHealth() == Active {
		node.Status = Active
		node.DataVersionId = node.GetDataVersion()
	}
	return node
}

func (n *Slave) Start() {
	nodePort := fmt.Sprintf("%d", n.Port)
	masterPort := fmt.Sprintf("%d", n.master.MasterPort)

	cmd := exec.Command(n.binary, masterPort, nodePort)

	// Daemon process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		fmt.Println("Error starting the process:", err)
		return
	}

	n.Status = Active
	n.ProcessId = cmd.Process.Pid

	go func() {
		err = cmd.Wait()

		if err != nil {
			fmt.Println("Process finished with error:", err)
			return
		}

		fmt.Println("Process finished successfully.")
	}()
}

func (n *Slave) CheckHealth() NodeStatus {
	resp, err := http.Get(n.healthURL)
	if err != nil {
		return Unrecoverable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return Active
	}

	return Unrecoverable

}

func (n *Slave) GetDataVersion() int64 {
	resp, err := http.Get(n.dataVersionURL)
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read response body: %v", err)
		}

		vid := string(body)
		if numberVal, err := strconv.ParseInt(vid, 10, 64); err == nil {
			return numberVal
		}
	}

	return -1

}

func (n *Slave) Broadcast(version int64) error {
	if version > n.DataVersionId {
		n.DataQuality = Dirty
		resp, err := http.Post(n.broadcastURL, "text/plain", nil)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			n.DataQuality = Fresh
			n.DataVersionId = version
		}
	}
	return nil
}

func (n *Slave) Shutdown() error {
	resp, err := http.Post(n.killURL, "text/plain", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New(resp.Status)

}

func (n *Slave) GetData() (*model.DataPayload, error) {
	resp, err := http.Get(fmt.Sprintf(n.dataUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result model.DataPayload
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, errors.New("No data found")
	}

	return &result, nil

}

func (n *Slave) Refresh(dataVersion int64) {
	data, err := n.GetData()
	if err != nil {
		fmt.Println("Error getting data:", err)
	}
	fmt.Println(fmt.Sprintf("Node Refresh: Port: %d, Node data version: %d, Master data version: %d", n.Port, n.DataVersionId, dataVersion))
	if data != nil {
		n.DataVersionId = data.DataVersion
		if dataVersion != n.DataVersionId {
			n.DataQuality = Dirty
		}
		n.ProcessId = data.PID
		n.RunningSince = data.RunningSince
	}
}
