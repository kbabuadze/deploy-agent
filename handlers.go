package main

import (
	"fmt"
	"log"
	"net/http"

	bolt "go.etcd.io/bbolt"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

// Recieves create signal, Unmarshals JSON, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleCreate(db *bolt.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var containerConfig ContainerConfig
		var deployment = Deployment{}

		if err := c.BindJSON(&containerConfig); err != nil {
			panic(err)
		}

		deployment.get(db, containerConfig.Name)

		if deployment.Name != "" {
			fmt.Println("deployment already exists")
			return
		}

		deployment = Deployment{
			Name:    containerConfig.Name,
			Config:  containerConfig,
			Running: make(map[string]container.ContainerCreateCreatedBody),
		}

		deployment.save(db)

		deployment.run(db)

		c.JSON(http.StatusOK, containerConfig)
	}
}

func handleStopDeploy(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		stopReq := struct {
			Name string `json:"name"`
		}{}

		if err := c.BindJSON(&stopReq); err != nil {
			panic(err)
		}

		deployment := Deployment{}

		deployment.get(db, stopReq.Name)

		if deployment.Name == "" {
			c.JSON(http.StatusNotFound, stopReq.Name)
			fmt.Printf("No such deployment")
			return
		}

		deployment.stop(db)

		c.JSON(http.StatusOK, deployment)
	}
}

func handleReset(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		db.Update(func(t *bolt.Tx) error {
			b := t.Bucket([]byte("Deployemnets"))
			b.Delete([]byte("nginx"))
			return nil
		})

		c.JSON(http.StatusOK, gin.H{
			"status": "reset",
		})
	}
}

func handleUpdate(db *bolt.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		updateParams := struct {
			Name  string `json:"name"`
			Image string `json:"image"`
		}{}

		if err := c.BindJSON(&updateParams); err != nil {
			if err != nil {
				log.Println(err.Error())
				return
			}

		}

		deployment := Deployment{}

		deployment.get(db, updateParams.Name)

		if deployment.Name == "" {
			fmt.Println("Could not find Deployment")
		}

		err := deployment.update(updateParams.Image, db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Failed to update, please check logs",
			})
			return
		}

	}
}

func handleGet(c *gin.Context) {

	result := []gin.H{}
	containers, err := GetContainers()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error retrieving containers",
		})
	}

	for _, container := range containers {
		result = append(result, gin.H{
			"id":     container.ID,
			"image":  container.Image,
			"name":   container.Names,
			"status": container.Status,
			"ports":  container.Ports,
		})
	}

	c.JSON(http.StatusOK, result)
}

func handleStatus(c *gin.Context) {

	type status struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	stat := make([]status, 0)

	containers, err := GetContainers()
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error retrieving containers",
		})
	}

	for _, container := range containers {
		stat = append(stat, status{
			ID:     container.ID,
			Status: container.Status,
		})
	}

	c.JSON(http.StatusOK, stat)
}
