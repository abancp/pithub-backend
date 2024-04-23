package main

import (
	"log"
	"net/http"
	"pithub-backend/auth"
	"pithub-backend/config"
	"pithub-backend/repo"
)

func main() {
	config.ConnectDb()

	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/signup", auth.Signup)
	http.HandleFunc("/repo/new", repo.CreateRepo)
	http.HandleFunc("/repo/checkname", repo.CheckName)
	http.HandleFunc("/repo/", repo.GetRepo)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
