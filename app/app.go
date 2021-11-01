package app

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kbabuadze/deploy-agent/domain"
	"github.com/kbabuadze/deploy-agent/svcs"
	bolt "go.etcd.io/bbolt"
)

type Agent struct {
	BasicAuthUser string
	BasicAuthPass string
	Port          string
	DBName        string
}

func Run(a *Agent) {

	listen_on := a.Port

	username := a.BasicAuthUser
	password := a.BasicAuthPass

	// Initialize gin
	r := gin.Default()

	// Open BoltDB
	db, err := bolt.Open(a.DBName, 0600, nil)
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

	deploymentRepo := domain.NewDeploymentsRepositoryDB(db)
	ctx := context.Background()
	deploymentRuntime := domain.NewDeploymentsRuntime(&ctx)

	deploymentHandler := DeploymentHandler{svcs.NewDeploymentService(&deploymentRepo, &deploymentRuntime)}

	// Setup Routes

	authorized.GET("/get/:name", deploymentHandler.GetDeployment)
	authorized.POST("/create", deploymentHandler.CreateDeployment)
	authorized.POST("/stop", deploymentHandler.StopDeployment)
	authorized.PATCH("/update", deploymentHandler.UpdateDeployment)

	// Start server
	r.Run(listen_on)
}
