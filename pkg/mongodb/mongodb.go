package mongodb

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Declare package-level variables
var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce sync.Once
func init(){
	clientInstance = DBinstance()
}
// DBinstance returns a singleton MongoDB client instance
func DBinstance() *mongo.Client {
	mongoOnce.Do(func() {
		// Initialize the MongoDB client
		MongoDb := "mongodb://localhost:27017"
		fmt.Println("MongoDB URI: ", MongoDb)

		clientOptions := options.Client().ApplyURI(MongoDb)
		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			clientInstanceError = err
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			clientInstanceError = err
			return
		}

		fmt.Println("Connected to MongoDB")
		clientInstance = client
	})

	if clientInstanceError != nil {
		log.Fatal(clientInstanceError)
	}
	return clientInstance
}

// OpenCollection opens a collection with the given name
func OpenCollection(collectionName string) *mongo.Collection {
	client := DBinstance()
	if client == nil {
		log.Fatal("Failed to create MongoDB client instance")
	}
	return client.Database("chat-app").Collection(collectionName)
}
