package main

import (
	"encoding/json"
	_ "io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBroadcastHandler(t *testing.T) {
	masterData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	masterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/data" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(masterData)
		} else {
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer masterServer.Close()

	req := httptest.NewRequest(http.MethodPost, "/notify", nil)
	w := httptest.NewRecorder()

	broadcastHandler(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if len(data) != len(masterData) {
		t.Errorf("expected data to have %d entries, but got %d", len(masterData), len(data))
	}

	for key, expectedValue := range masterData {
		if value, ok := data[key]; !ok || value != expectedValue {
			t.Errorf("expected data[%q] = %q, but got %q", key, expectedValue, value)
		}
	}
}
