package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Real-time Multiplayer Game Server")
	fmt.Println("Starting server on :8080...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
