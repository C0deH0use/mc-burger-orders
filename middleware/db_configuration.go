package middleware

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mc-burger-orders/log"
	"os"
	"time"
)

func GetMongoClient() *mongo.Database {
	db, exists := os.LookupEnv("DB_MONGO_NAME")
	if !exists {
		log.Error.Println("DB_MONGO_NAME is empty, running the default")
		db = "order-db"
	}
	opts := getConnectionOptions()

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Error.Fatalln(err)
	}

	database := client.Database(db)
	return database
}

func getConnectionOptions() *options.ClientOptions {
	url := os.Getenv("DB_MONGO_URL")
	if len(url) == 0 {
		log.Error.Fatalln("MONGO DB URL is missing!")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	return options.Client().
		ApplyURI(url).
		SetAuth(getConnectionCredentials()).
		SetConnectTimeout(60 * time.Second).
		SetMaxPoolSize(10).
		SetMinPoolSize(2).
		SetRetryReads(true).
		SetRetryWrites(true).
		SetServerAPIOptions(serverAPI)
}

func getConnectionCredentials() options.Credential {
	user := os.Getenv("DB_MONGO_USER")
	password := os.Getenv("DB_MONGO_PASSWORD")
	if len(user) == 0 || len(password) == 0 {
		log.Error.Panicf("Mongo User\\Password are not specified")
	}

	return options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	}
}
