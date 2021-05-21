package main

import (
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

type ContainerConfig struct {
	Image         string   `json:"image"`
	Name          string   `json:"name"`
	ContainerPort string   `json:"containerPort"`
	HostIP        string   `json:"hostIP"`
	HostPortFirst int      `json:"hostPortFirst"`
	Replicas      int      `json:"replicas"`
	Command       []string `json:"command"`
}

var containers = map[string]container.ContainerCreateCreatedBody{}

func main() {

	// Initialize gin
	r := gin.Default()

	// Setup Routes
	r.POST("/create", handleCreate)

	r.GET("/stop", handleStop)

	r.GET("/get", handleGet)

	r.GET("/status", handleStatus)

	r.POST("/update", handleUpdate)

	// Start server
	r.Run("0.0.0.0:8008")
}
