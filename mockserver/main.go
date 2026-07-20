// mockserver/main.go
//
// A minimal stand-in for the external summary API.
// Run it with:  go run ./mockserver
//
// It listens on :9090 and always responds with a realistic
// fake summary for whatever DB credentials it receives.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type summaryRequest struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	User   string `json:"user"`
	DBName string `json:"dbname"`
}

type tableResult struct {
	Name     string  `json:"name"`
	RowCount int64   `json:"row_count"`
	SizeMB   float64 `json:"size_mb"`
}

type schemaResult struct {
	Name   string        `json:"name"`
	Tables []tableResult `json:"tables"`
}

type summaryResponse struct {
	SummaryID string         `json:"summary_id"`
	Schemas   []schemaResult `json:"schemas"`
}

func main() {
	http.HandleFunc("/api/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req summaryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		log.Printf("[mock] received request → host=%s dbname=%s user=%s",
			req.Host, req.DBName, req.User)

		resp := summaryResponse{
			SummaryID: fmt.Sprintf("mock-sum-%s", req.DBName),
			Schemas: []schemaResult{
				{
					Name: "public",
					Tables: []tableResult{
						{Name: "users", RowCount: 1240, SizeMB: 12.5},
						{Name: "orders", RowCount: 580, SizeMB: 8.3},
						{Name: "products", RowCount: 320, SizeMB: 4.1},
					},
				},
				{
					Name: "analytics",
					Tables: []tableResult{
						{Name: "events", RowCount: 95000, SizeMB: 210.7},
						{Name: "sessions", RowCount: 12400, SizeMB: 55.2},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)

		log.Printf("[mock] responded with summary_id=%s", resp.SummaryID)
	})

	// Health-check so you can verify the mock is up.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status":"ok"}`)
	})

	addr := ":9090"
	log.Printf("[mock] external service listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
