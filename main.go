package main

import (
	"fmt"
	"log"

	bolt "go.etcd.io/bbolt"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

var containers = map[string]container.ContainerCreateCreatedBody{}

func main() {

	// Initialize gin
	r := gin.Default()

	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Deployments"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	// Setup Routes
	r.POST("/create", handleCreate(db))

	r.GET("/stop", handleStop)

	r.POST("/stop", handleStopDeploy(db))

	r.GET("/get", handleGet)

	r.GET("/reset", handleReset(db))

	r.GET("/status", handleStatus)

	r.POST("/update", handleUpdate)

	// Start server
	r.Run("0.0.0.0:8008")
}
