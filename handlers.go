package main

import (
	"net/http"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

func errorAndExit(c *gin.Context, err string, status int, message string) {
	os.Stderr.WriteString("[Deploy Agent Error] " + time.Now().String() + " " + err + " \n")
	c.JSON(status, gin.H{
		"message": message,
	})
}

func successAndExit(c *gin.Context, status int, message string) {
	os.Stdout.WriteString("[Deploy Agent Success] " + time.Now().String() + " " + message)
	c.JSON(status, gin.H{
		"message": message,
	})
}

// Recieves create signal, Unmarshals JSON, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleCreate(db *bolt.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var containerConfig ContainerConfig
		var deployment = Deployment{}

		if err := c.ShouldBindJSON(&containerConfig); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
			return
		}

		deployment.get(db, containerConfig.Name)

		if deployment.Name != "" {
			errorAndExit(c, "deployment not found", http.StatusNotFound, "deployment Not Found")
			return
		}

		deployment = Deployment{
			Name:    containerConfig.Name,
			Config:  containerConfig,
			Running: make(map[string]container.ContainerCreateCreatedBody),
		}

		if err := deployment.save(db); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error saving deploy, please check logs")
			return
		}

		if err := deployment.run(db); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "failed to run deployment, please check logs")
		}

		successAndExit(c, http.StatusCreated, "deployment"+deployment.Name+"successfuly created")
	}
}

func handleStopDeploy(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		stopReq := struct {
			Name string `json:"name"`
		}{}

		if err := c.ShouldBindJSON(&stopReq); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
			return
		}

		deployment := Deployment{}

		deployment.get(db, stopReq.Name)

		if deployment.Name == "" {
			c.JSON(http.StatusNotFound, stopReq.Name)
			errorAndExit(c, "deployment not found", http.StatusNotFound, "deployment not found")
			return
		}

		if err := deployment.stop(db); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error stopping deployment, please check logs")
			return
		}

		successAndExit(c, http.StatusOK, "deployment successfully stopped")
	}
}

func handleUpdate(db *bolt.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		updateParams := struct {
			Name  string `json:"name"`
			Image string `json:"image"`
		}{}

		if err := c.ShouldBindJSON(&updateParams); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
			return
		}

		deployment := Deployment{}

		if err := deployment.get(db, updateParams.Name); err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error getting deployment, please check logs")
			return
		}

		if deployment.Name == "" {
			errorAndExit(c, "deployment not found", http.StatusNotFound, "deployment not found")
			return
		}

		err := deployment.update(updateParams.Image, db)
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "update failed, please check the logs")
			return
		}

	}
}

func handleGet(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		deployment := Deployment{}

		deployment.get(db, name)

		result := []gin.H{}
		containers, err := GetContainers()
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error retrieving containers, please check logs")
			return
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
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error retrieving containers, please check logs")
		return
	}

	for _, container := range containers {
		stat = append(stat, status{
			ID:     container.ID,
			Status: container.Status,
		})
	}

	c.JSON(http.StatusOK, stat)
}
