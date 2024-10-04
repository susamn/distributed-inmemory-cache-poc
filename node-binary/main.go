package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var node *Node

func nodeDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	finalResponse, err := json.Marshal(DataPayload{DataVersion: node.DataVersion, Data: node.Data, PID: node.PID, RunningSince: node.RunningSince})

	if err != nil {
		http.Error(w, "Failed to marshal map to JSON", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(finalResponse)
	if err != nil {
		return
	}
}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/replicate/data", node.MasterPort))
	if err != nil {
		http.Error(w, "Failed to consume master API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	var result DataPayload
	if err := json.Unmarshal(body, &result); err != nil {
		http.Error(w, "Failed to unmarshal response", http.StatusInternalServerError)
		return
	}

	node.Data = result.Data
	node.DataVersion = result.DataVersion
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <master-port> <node-port>")
	}

	masterServicePort, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid master port number: %v", err)
	}

	nodePort, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatalf("Invalid node port number: %v", err)
	}

	shutdownChan := make(chan bool, 1)

	pid := os.Getpid()

	node = NewNode(nodePort, masterServicePort, shutdownChan, pid)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", nodePort),
	}

	http.HandleFunc("/data", nodeDataHandler)
	http.HandleFunc("/dataVersion", nodeDataVersionHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/notify", broadcastHandler)
	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Shutting down node..."))
		shutdownChan <- true
	})

	go func() {
		fmt.Printf("Node running on port %d\n", nodePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %d: %v\n", nodePort, err)
		}
	}()

	<-shutdownChan

	fmt.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	fmt.Println("Server exited cleanly")

}

func nodeDataVersionHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(fmt.Sprintf("%d", node.DataVersion)))
}

func healthHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("OK"))
}
