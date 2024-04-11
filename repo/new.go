package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pithub-backend/config"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type ReqRepo struct {
	Name        string   `json:"name"`
	Secure      string   `json:"secure"`
	Description string   `json:"description"`
	CodeURL     string   `json:"codeURL"`
	Languages   []string `json:"languages"`
	LiveURL     string   `json:"liveURL"`
}

type Repo struct {
	Name        string
	Secure      string
	Description string
	CodeURL     string
	Languages   []string
	LiveURL     string
	Date        int64
	Username    string
}

func CreateRepo(w http.ResponseWriter, r *http.Request) {
	//TODO:Middleware for token validation
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method != "POST" {
		http.Error(w, "Method not accepted", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error in parsing form!", http.StatusNotFound)
		return
	}

	tokenCookie, err := r.Cookie("token")

	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tokenString := tokenCookie.Value
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("JWT_SECRET_KEY"), nil
	})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}
	if !token.Valid {
		http.Error(w, "Invalid token , unauthorized", http.StatusUnauthorized)
		return
	}
	var reqRepo ReqRepo
	err = json.NewDecoder(r.Body).Decode(&reqRepo)
	if err != nil {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error parsing claims")
		return
	}
	username, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	time := time.Now().UnixMilli()

	repo := Repo{reqRepo.Name, reqRepo.Secure, reqRepo.Description, reqRepo.CodeURL, reqRepo.Languages, reqRepo.LiveURL, time, username}
	
	db := config.DB

	_, err = db.Collection("repos").InsertOne(context.Background(), repo)

	if err != nil {
		http.Error(w, "Something Went Wrong ", http.StatusInternalServerError)
		return
	}
	http.Error(w,"Something Went Wrong!",500)
}
