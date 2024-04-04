package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Define a handler function to respond to requests
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World! You've reached the server on port 8000\n")
	})

	// Start the server on port 8000
	err := http.ListenAndServe(":8000", nil)

	// Handle potential errors
	if err != nil {
		panic(err)
	}
}
