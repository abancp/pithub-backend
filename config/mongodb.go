package config

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectDb() {
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017/")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		panic(err)
	}

	// Ping the MongoDB server to check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to MongoDB!")
	DB = client.Database("pithub")
}
