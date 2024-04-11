package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pithub-backend/config"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"name"`
	Username string             `bson:"username"`
	Email    string             `bson:"email"`
	Password []byte             `bson:"password"`
}

type ReqUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ReqLoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func Login(w http.ResponseWriter, r *http.Request) {
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

	var reqBody ReqLoginUser
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Something Went Wrong!", http.StatusInternalServerError)
		return
	}

	email := reqBody.Email
	password := reqBody.Password
	fmt.Println(email, password)

	db := config.DB

	var user User

	err = db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No matching document found")
			return
		}
		http.Error(w, "Something Went Wrong!", 500)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			http.Error(w, "User not found or Password not Matching", http.StatusUnauthorized)
		} else {
			http.Error(w, "Failed to compare passwords", http.StatusInternalServerError)
		}
		return
	} else {
		var jwtKey = []byte("JWT_SECRET_KEY")
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["authorized"] = true
		claims["username"] = user.Username
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error generating token")
			return
		}
		responseData := Response{
			Success: true,
			Message: "Login successful",
		}

		responseJSON, err := json.Marshal(responseData)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
		cookie := http.Cookie{
			Name:     "token",
			Value:    tokenString,
			HttpOnly: true,
			Expires:  time.Now().Add(24 * time.Hour),
		}

		// Set the cookie in the response
		http.SetCookie(w, &cookie)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not accepted!", http.StatusNotFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error in parsing form!", http.StatusNotFound)
		return
	}

	var reqUser ReqUser
	err := json.NewDecoder(r.Body).Decode(&reqUser)
	if err != nil {
		panic(err)
	}
	hashedPassword, err1 := bcrypt.GenerateFromPassword([]byte(reqUser.Password), bcrypt.DefaultCost)
	if err1 != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	db := config.DB
	user := bson.D{
		primitive.E{Key: "name", Value: reqUser.Name},
		primitive.E{Key: "username", Value: reqUser.Username},
		primitive.E{Key: "email", Value: reqUser.Email},
		primitive.E{Key: "password", Value: hashedPassword},
	}

	var existUser User

	err2 := db.Collection("users").FindOne(context.Background(), bson.M{"email": reqUser.Email}).Decode(&existUser)
	if err2 != nil {
		if err2 != mongo.ErrNoDocuments {
			log.Fatal(err2)
			return
		}
	} else {
		responseData := Response{
			Success: false,
			Message: "email:email already used",
		}

		responseJSON, err5 := json.Marshal(responseData)
		if err5 != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(responseJSON)
		return
	}

	err3 := db.Collection("users").FindOne(context.Background(), bson.M{"username": reqUser.Username}).Decode(&existUser)
	if err3 != nil {
		if err3 != mongo.ErrNoDocuments {
			log.Fatal(err3)
			return
		}
	} else {
		responseData := Response{
			Success: false,
			Message: "username:username already used",
		}

		responseJSON, err5 := json.Marshal(responseData)
		if err5 != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(responseJSON)
		return
	}

	result, err4 := db.Collection("users").InsertOne(context.Background(), user)
	fmt.Println(result)
	if err4 != nil {
		log.Fatal("Error finding document:", err4)
		return
	}
	var jwtKey = []byte("JWT_SECRET_KEY")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["username"] = reqUser.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, err4 := token.SignedString(jwtKey)
	if err4 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Error generating token")
		return
	}
	responseData := Response{
		Success: true,
		Message: "User created successful",
	}

	responseJSON, err5 := json.Marshal(responseData)
	if err5 != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	}

	// Set the cookie in the response
	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
