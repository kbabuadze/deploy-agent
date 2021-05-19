package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

// Recieves create signal, Unmarshals JSON, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleCreate(c *gin.Context) {

	var containerConfig ContainerConfig
	var containerBody container.ContainerCreateCreatedBody
	var err error

	if err := c.BindJSON(&containerConfig); err != nil {
		panic(err)
	}

	go func() {
		for i := 0; i < containerConfig.Replicas; i++ {

			containerProps := ContainerProps{
				Image:    containerConfig.Image,
				Name:     containerConfig.Name + "-" + fmt.Sprint(i+1),
				Port:     containerConfig.ContainerPort + "/tcp",
				HostIP:   containerConfig.HostIP,
				HostPort: fmt.Sprint(containerConfig.HostPortFirst+i) + "/tcp",
				Command:  containerConfig.Command,
				Label:    map[string]string{"by": "deploy-agent"},
			}

			if containerBody, err = DeployContainer(containerProps); err != nil {
				log.Println(err.Error())
				return
			}

			containers = append(containers, containerBody)
		}

	}()

	c.JSON(http.StatusOK, gin.H{"status": "recieved create signal, please check /status for more details"})

}

// Recieves stop signal, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleStop(c *gin.Context) {

	go func() {
		for _, container := range containers {
			err := StopContainer(container.ID, 60*time.Second)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"status": "failed",
				})
			}
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"status": "received stop signal, please check /status for more details",
	})
}

// Lists
func handleGet(c *gin.Context) {

	result := []gin.H{}
	containers, err := GetContainerStatus()
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

	containers, err := GetContainerStatus()
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
