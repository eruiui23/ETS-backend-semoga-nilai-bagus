package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	storage := NewStorage()
	handler := NewHandler(storage)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	port := ":8080"
	fmt.Printf("Server PinjamKampus berjalan di http://localhost%s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
