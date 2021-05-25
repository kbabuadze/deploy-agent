package main

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var containers = map[string]container.ContainerCreateCreatedBody{}

func main() {

	//Setup .env
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	listen_on := os.Getenv("LISTEN_ON")

	// Initialize gin
	r := gin.Default()

	// Open BoltDB
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize Bucket if it does not exist
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Deployments"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	// Setup Routes
	r.POST("/create", handleCreate(db))

	r.POST("/stop", handleStopDeploy(db))

	r.GET("/get", handleGet)

	r.GET("/reset", handleReset(db))

	r.GET("/status", handleStatus)

	r.POST("/update", handleUpdate)

	// Start server
	r.Run(listen_on)
}
