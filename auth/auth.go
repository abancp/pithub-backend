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

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Method not accepted", http.StatusNotFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error in parsing form!", http.StatusNotFound)
		return
	}
	email, password := r.FormValue("email"), r.FormValue("password")

	db := config.DB

	var user User

	err := db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Println("No matching document found")
			return
		}
		log.Fatal(err)
	}

	err1 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err1 != nil {
		if err1 == bcrypt.ErrMismatchedHashAndPassword {
			http.Error(w, "User not found or Password not Matching", http.StatusUnauthorized)
		} else {
			http.Error(w, "Failed to compare passwords", http.StatusInternalServerError)
		}
		return
	} else {
		var jwtKey = []byte("its_my_secret_key_of_passwod_of_kuntham")
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

		responseJSON, err1 := json.Marshal(responseData)
		if err1 != nil {
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

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not accepted!", http.StatusNotFound)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error in parsing form!", http.StatusNotFound)
		return
	}

	name, username, email, password := r.FormValue("name"), r.FormValue("username"), r.FormValue("email"), r.FormValue("password")
	hashedPassword, err1 := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err1 != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	db := config.DB
	user := bson.D{
		primitive.E{Key: "name", Value: name},
		primitive.E{Key: "username", Value: username},
		primitive.E{Key: "email", Value: email},
		primitive.E{Key: "password", Value: hashedPassword},
	}

	var existUser User

	err2 := db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&existUser)
	if err2 != nil {
		if err2 != mongo.ErrNoDocuments {
			log.Fatal(err2)
			return
		}
	} else {
		responseData := Response{
			Success: false,
			Message: "email already used",
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

	err3 := db.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(&existUser)
	if err3 != nil {
		if err3 != mongo.ErrNoDocuments {
			log.Fatal(err3)
			return
		}
	} else {
		responseData := Response{
			Success: false,
			Message: "username already used",
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
	var jwtKey = []byte("its_my_secret_key_of_passwod_of_kuntham")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["username"] = username
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
