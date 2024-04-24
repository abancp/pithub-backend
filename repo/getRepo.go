package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pithub-backend/config"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbGetRepo struct {
	Id          string   `bson:"_id"`
	Name        string   `bson:"name"`
	Username    string   `bson:"username"`
	Description string   `bson:"description"`
	Secure      string   `bson:"secure"`
	CodeURL     string   `bson:"codeURL"`
	LiveURL     string   `bson:"liveURL"`
	Languages   []string `bson:"languages"`
	Date        int64    `bson:"date"`
}

type ResRepo struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Description string   `json:"description"`
	Secure      string   `json:"secure"`
	CodeURL     string   `json:"codeURL"`
	LiveURL     string   `json:"liveURL"`
	Languages   []string `json:"languages"`
	Date        int64    `json:"date"`
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not accepted", http.StatusNotFound)
		return
	}
	params := strings.Split(r.URL.Path[len("/repo/"):], "/")
	if len(params) != 2 {
		http.Error(w, "Not found!", http.StatusNotFound)
		return
	}
	username := params[0]
	reponame := params[1]

	tokenCookie, err := r.Cookie("token")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		http.Error(w, "Invalid token , Unauthorized", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error parsing claims")
		return
	}
	tokenUsername, ok := claims["username"].(string)
	if !ok {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	owner := false 
	if tokenUsername == username {
		owner = true
	}
	fmt.Println(owner)


	db := config.DB
	var repo DbGetRepo
	err = db.Collection("repos").FindOne(context.Background(), bson.M{"username": username, "name": reponame}).Decode(&repo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			errorResponse := map[string]string{"success": "false", "message": "No repository found"}
			responseJSON, err := json.Marshal(errorResponse)
			if err != nil {
				http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(responseJSON)
			return
		}
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	resData := ResRepo(repo)
	responseJSON, err := json.Marshal((resData))
	if err != nil {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
