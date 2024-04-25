package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pithub-backend/auth"
	"pithub-backend/config"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type NewRepoReqBody struct {
	Name        string   `json:"name"`
	Secure      string   `json:"secure"`
	Description string   `json:"description"`
	CodeURL     string   `json:"codeURL"`
	Languages   []string `json:"languages"`
	LiveURL     string   `json:"liveURL"`
	Token       string   `json:"token"`
}

type Repo struct {
	Name        string   `bson:"name"`
	Secure      string   `bson:"secure"`
	Description string   `bson:"description"`
	CodeURL     string   `bson:"codeURL"`
	Languages   []string `bson:"languages"`
	LiveURL     string   `bson:"liveURL"`
	Date        int64    `bson:"date"`
	Username    string   `bson:"username"`
	Tasks 		Tasks    `bson:"tasks"`
}

func CreateRepo(w http.ResponseWriter, r *http.Request) {
	//TODO:Middleware for token validation
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

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

	var reqBody NewRepoReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	tokenString := reqBody.Token
	fmt.Println(tokenString)
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("JWT_SECRET_KEY"), nil
	})
	if err != nil {
		fmt.Println("Error parsing token:", err)
		http.Error(w, "Something Went Wrong!"+err.Error(), http.StatusInternalServerError)
		return
	}
	if !token.Valid {
		http.Error(w, "Invalid token , unauthorized", http.StatusUnauthorized)
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

	repo := Repo{reqBody.Name, reqBody.Secure, reqBody.Description, reqBody.CodeURL, reqBody.Languages, reqBody.LiveURL, time, username,Tasks{ActiveTasks: []string{},WorkingTasks: []string{},ClosedTasks: []string{}}}

	db := config.DB

	_, err = db.Collection("repos").InsertOne(context.Background(), repo)

	if err != nil {
		http.Error(w, "Something Went Wrong "+err.Error(), http.StatusInternalServerError)
		return
	}
	responseData := auth.Response{
		Success: true,
		Message: "Repository Created Successfully",
	}

	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
