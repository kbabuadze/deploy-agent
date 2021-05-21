package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
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
				Name:     containerConfig.Name + "-" + strconv.Itoa(i+1),
				Port:     containerConfig.ContainerNet.Port + "/" + containerConfig.ContainerNet.Proto,
				HostIP:   containerConfig.HostNet.IP,
				HostPort: strconv.Itoa(containerConfig.HostNet.PortFirst+i) + "/" + containerConfig.HostNet.Proto,
				Command:  containerConfig.Command,
				Label:    map[string]string{"by": "deploy-agent"},
			}

			if containerBody, err = DeployContainer(containerProps); err != nil {
				log.Println(err.Error())
				return
			}

			containers[containerBody.ID] = containerBody
		}
	}()
	fmt.Printf("%v", containerConfig)
	c.JSON(http.StatusOK, containerConfig)

}

// Recieves stop signal, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleStop(c *gin.Context) {

	fmt.Println(containers)
	go func() {
		for _, container := range containers {
			err := StopContainer(container.ID, 60*time.Second)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if err := RemoveContainer(container.ID); err != nil {
				fmt.Println(err.Error())
				return
			}

			delete(containers, container.ID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"status": "received stop signal, please check /status for more details",
	})
}

type UpdateContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func handleUpdate(c *gin.Context) {

	containers, err := GetContainers()
	if err != nil {
		log.Println(err.Error())
		return
	}

	updateParams := UpdateContainer{}

	if err := c.BindJSON(&updateParams); err != nil {
		if err != nil {
			log.Println(err.Error())
			return
		}

	}

	updateSlice := make([]types.Container, 0)

	for _, container := range containers {
		if true {
			fmt.Printf(container.Names[0])
			updateSlice = append(updateSlice, container)
		}
	}

	go func() {
		_ = UpdateContainers(updateParams.Name, updateParams.Image, updateSlice)
	}()

	c.JSON(http.StatusOK, updateParams)

}

// List Continers
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
