package main

import (
	c "distributed-inmemory-cache/config"
	"distributed-inmemory-cache/engine"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var master *engine.Master
var conf *c.Config

func main() {
	var err error
	conf, err = c.ReadConfig()
	if err != nil {
		log.Fatal("Error reading config: ", err)
		os.Exit(-1)
	}

	master = engine.NewMaster(conf)
	master.MakeAvailable()
	http.HandleFunc("/replicate/data", replicateDataHandler)
	http.HandleFunc("/api/data/get", getDataHandler)
	http.HandleFunc("/api/data/set", setDataHandler)
	http.HandleFunc("/api/data/delete", deleteDataHandler)
	http.HandleFunc("/api/infra/scaleup", infraScaleUpHandler)
	http.HandleFunc("/api/infra/scaledown", infraScaleDownHandler)
	http.HandleFunc("/api/infra/killall", killAllHandler)
	http.HandleFunc("/api/infra/nodestats", nodeCountHandler)

	fs := http.FileServer(http.Dir("public"))

	http.Handle("/", fs)

	http.Handle("/css/", fs)
	http.Handle("/js/", fs)

	addr := fmt.Sprintf(":%d", conf.Service.Master.Port)
	fmt.Printf("Server running on port %d and server is ready !!!\n", conf.Service.Master.Port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func killAllHandler(w http.ResponseWriter, request *http.Request) {
	err := master.KillAllNodes()
	if err != nil {
		log.Println("Error killing all nodes: ", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}
	w.WriteHeader(http.StatusOK)

}

func replicateDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	finalResponse, err := json.Marshal(master.GetReplicationData())

	log.Println("Replication API: Get replication data called")

	if err != nil {
		http.Error(w, "Failed to marshal map to JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(finalResponse)
	if err != nil {
		return
	}
}

func nodeCountHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	finalResponse, err := json.Marshal(master.NodeStats())

	if err != nil {
		http.Error(w, "Failed to marshal map to JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(finalResponse)
	if err != nil {
		return
	}
}

func infraScaleDownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Infra API: Scale DOWN called")
	state := master.ScaleDown(conf)
	if state {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func infraScaleUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Infra API: Scale UP called")
	state := master.ScaleUp(conf)
	if state {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func getDataHandler(w http.ResponseWriter, request *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	finalResponse, err := json.Marshal(master.GetData())

	if err != nil {
		http.Error(w, "Failed to marshal map to JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(finalResponse)
	if err != nil {
		return
	}
}

func setDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Data API: Set called")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data map[string]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Println("Received data:", data)

	master.SetData(data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(master.GetData())
}

func deleteDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Println("Data API: Delete called")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var data []string
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Println("Received data:", data)

	master.DeleteData(data)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(master.GetData())
}

//TIP See GoLand help at <a href="https://www.jetbrains.com/help/go/">jetbrains.com/help/go/</a>.
// Also, you can try interactive lessons for GoLand by selecting 'Help | Learn IDE Features' from the main menu.
