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

	// Bootstrap container

	// Receive command though http

	r := gin.Default()

	// Create unmarshals recieved JSON and gives it to go routine
	// further info can be obtained through /status
	r.POST("/create", handleCreate)

	r.GET("/stop", handleStop)

	r.GET("/get", handleGet)

	r.GET("/status", handleStatus)

	r.POST("/update", handleUpdate)

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	// Download container

	// Tag Container

	// Run container

	// Monitor Container

	// Send email when container is up and if it fails

	r.Run("0.0.0.0:8008")
}
