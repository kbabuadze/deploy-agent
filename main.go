package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

type ContainerConfig struct {
	Image         string   `json:"image"`
	ContainerPort string   `json:"containerPort"`
	HostIP        string   `json:"hostIP"`
	HostPortFirst int      `json:"hostPortFirst"`
	Replicas      int      `json:"replicas"`
	Command       []string `json:"command"`
}

func main() {

	// Bootstrap container

	containers := make([]container.ContainerCreateCreatedBody, 0)

	// Receive command though http

	r := gin.Default()

	r.POST("/create", func(c *gin.Context) {

		var containerConfig ContainerConfig
		var containerBody container.ContainerCreateCreatedBody
		var err error

		if err := c.BindJSON(&containerConfig); err != nil {
			panic(err)
		}

		for i := 0; i < containerConfig.Replicas; i++ {

			containerProps := ContainerProps{
				Image:    containerConfig.Image,
				Port:     containerConfig.ContainerPort + "/tcp",
				HostIP:   containerConfig.HostIP,
				HostPort: fmt.Sprint(containerConfig.HostPortFirst+i) + "/tcp",
				Command:  containerConfig.Command,
			}

			if containerBody, err = DeployContainer(containerProps); err != nil {
				panic(err)
			}

			containers = append(containers, containerBody)
		}

		c.JSON(http.StatusOK, containers)
	})

	r.GET("/stop", func(c *gin.Context) {

		for _, container := range containers {
			err := StopContainer(container.ID, 60*time.Second)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"status": "failed",
				})
			}

		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
		})
	})

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	// Download container

	// Tag Container

	// Run container

	// Monitor Container

	// Send email when container is up and if it fails

	r.Run("0.0.0.0:8008")
}
