package main

import (
	"log"
	"net/http"
	"pithub-backend/auth"
	"pithub-backend/config"
)

func main() {
	config.ConnectDb()
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/signup", auth.Register)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
