package main

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kbabuadze/deploy-agent/domain"
	"github.com/kbabuadze/deploy-agent/svcs"
)

// var containers = map[string]container.ContainerCreateCreatedBody{}

func main() {

	//Setup .env
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	listen_on := os.Getenv("LISTEN_ON")

	username := os.Getenv("DEPLOY_USERNAME")
	password := os.Getenv("DEPLOY_PASSWORD")

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

	// Setup basic auth
	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		username: password,
	}))

	deoloymentRepo := domain.NewDeploymentsRepositoryDB(db)

	deploymentHandler := DeploymentHandler{svcs.NewDploymentService(&deoloymentRepo)}

	// Setup Routes
	authorized.POST("/create", handleCreate(db))

	authorized.POST("/stop", handleStopDeploy(db))

	authorized.GET("/get", handleGet(db))

	authorized.GET("/status", handleStatus)

	authorized.PATCH("/update", handleUpdate(db))

	authorized.GET("/abstractGet/:name", deploymentHandler.GetDeployments)

	// Start server
	r.Run(listen_on)
}
