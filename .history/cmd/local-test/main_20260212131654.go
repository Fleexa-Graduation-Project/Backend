package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"errors"
)



func main() {
	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event map[string]interface{}

		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}

		err := handler.HandleRequest(context.Background(), event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
