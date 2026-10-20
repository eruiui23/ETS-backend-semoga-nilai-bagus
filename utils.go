package main

import (
	"encoding/json"
	"net/http"
)

// writeJSON membantu mengirim response JSON beserta HTTP status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError mengirim format error standar dalam bentuk JSON.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
