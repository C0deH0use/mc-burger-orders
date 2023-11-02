package middleware

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
)

func GetMongoClient() *mongo.Database {
	u := os.Getenv("DB_MONGO_URL")
	if len(u) == 0 {
		panic("MONGO DB URL is missing!")
	}
	db, exists := os.LookupEnv("DB_MONGO_NAME")
	if !exists {
		log.Println("DB_MONGO_NAME is empty, running the default")
		db = "order-db"
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(u).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	database := client.Database(db)
	return database
}
