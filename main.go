package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	handler := NewHandler()

	port := ":8080"
	fmt.Printf("Server PinjamKampus berjalan di http://localhost%s\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
