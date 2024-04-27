package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"pithub-backend/auth"
	"pithub-backend/config"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DbRepo struct {
	Name        string   `bson:"name"`
	Secure      string   `bson:"secure"`
	Description string   `bson:"description"`
	CodeURL     string   `bson:"codeURL"`
	Languages   []string `bson:"languages"`
	LiveURL     string   `bson:"liveURL"`
	Date        int64    `bson:"date"`
	Username    string   `bson:"username"`
}

type ReqBody struct {
	Name string `json:"name"`
}

func CheckName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("CLIENT_URL"))
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With,Authorization, Content-Type, Accept")
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
	authHeader := r.Header.Get("Authorization")
	isBearer := strings.HasPrefix(authHeader, "Bearer ")
	if !isBearer {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// //-------------
	// cookies := make(map[string]*http.Cookie)
	// for _, cookie := range r.Cookies() {
	// 	cookies[cookie.Name] = cookie
	// }
	// fmt.Println(cookies)
	// tokenCookie, err := r.Cookie("token")

	// if err != nil {
	// 	fmt.Println(err)
	// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 	return
	// }
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("JWT_SECRET_KEY"), nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err,token)
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	if !token.Valid {
		http.Error(w, "Invalid token , unauthorized", http.StatusUnauthorized)
		return
	}
	//-------
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Error parsing claims")
		return
	}

	username, ok := claims["username"].(string)
	fmt.Println(username)
	if !ok {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	var reqBody ReqBody
	err = json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}
	name := reqBody.Name
	db := config.DB

	var repo DbRepo
	err = db.Collection("repos").FindOne(context.Background(), bson.M{"username": username, "name": name}).Decode(&repo)
	if err != nil {
		if err == mongo.ErrNoDocuments {

			responseData := auth.Response{
				Success: true,
				Message: "name is available",
			}

			responseJSON, err := json.Marshal(responseData)
			if err != nil {
				http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(responseJSON)
			return
		}
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}
	responseData := auth.Response{
		Success: false,
		Message: "name is unavailable",
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
