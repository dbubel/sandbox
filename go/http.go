package main

import (
	"fmt"
	"net/http"
)

// Handler function for the root route "/"
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, this is a response from the Go HTTP server!")
}

func main() {
	// Register the handler function for the root route
	http.HandleFunc("/", handler)

	// Specify the port to listen on (8080) and start the server
	fmt.Println("Server is listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
