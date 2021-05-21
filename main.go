package main

import (
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

type ContainerConfig struct {
	Image string `json:"image"`
	Name  string `json:"name"`

	ContainerNet struct {
		Port  string `json:"port"`
		Proto string `json:"proto"`
	} `json:"containerNet"`

	HostNet struct {
		IP        string `json:"ip"`
		PortFirst int    `json:"portFirst"`
		Proto     string `json:"proto"`
	} `json:"hostNet"`

	Replicas int      `json:"replicas"`
	Command  []string `json:"command"`
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
