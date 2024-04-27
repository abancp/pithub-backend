package auth

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

type ResUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not accepted", http.StatusNotFound)
		return
	}

	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("JWT_SECRET_KEY"), nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err, token)
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	if !token.Valid {
		http.Error(w, "Invalid token , unauthorized", http.StatusUnauthorized)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error parsing claims")
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	username, ok := claims["username"].(string)
	fmt.Println(username)
	if !ok {
		fmt.Println(ok,"here")
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	var user User
	db := config.DB
	err = db.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found , Unauthorized", http.StatusUnauthorized)
			return
		} else {
			fmt.Println(err)
			http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
			return
		}
	}

	resUser := ResUser{Username: user.Username, Email: user.Email, Fullname: user.Name}
	responseJSON, err := json.Marshal(resUser)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
