package middleware

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mc-burger-orders/log"
	"os"
)

func GetMongoClient() *mongo.Database {
	url := os.Getenv("DB_MONGO_URL")
	if len(url) == 0 {
		log.Error.Fatalln("MONGO DB URL is missing!")
	}
	db, exists := os.LookupEnv("DB_MONGO_NAME")
	if !exists {
		log.Error.Println("DB_MONGO_NAME is empty, running the default")
		db = "order-db"
	}
	user := os.Getenv("DB_MONGO_USER")
	password := os.Getenv("DB_MONGO_PASSWORD")

	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(url).SetAuth(auth).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Error.Fatalln(err)
	}

	database := client.Database(db)
	return database
}
