package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println("Starting Task Service...")

	// Konfigurasi akan ditambahkan di sini nanti (port, db conn string, dll)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Port default untuk task-service
	}

	// Setup router HTTP (akan menggunakan chi/gin nanti)
	// router := SetupRouter() // Fungsi ini akan dibuat nanti

	// Untuk sementara, buat handler sederhana
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Task Service is healthy!")
	})

	log.Printf("Task Service listening on port %s", port)
	// err := http.ListenAndServe(":"+port, router) // Akan diaktifkan nanti
	err := http.ListenAndServe(":"+port, nil) // Gunakan handler default sementara
	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}
}
