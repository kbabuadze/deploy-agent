package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/kbabuadze/deploy-agent/domain"
	"github.com/kbabuadze/deploy-agent/svcs"
)

type DeploymentHandler struct {
	service svcs.DeploymentService
}

func errorAndExit(c *gin.Context, err string, status int, message string) {
	os.Stderr.WriteString("[Deploy Agent Error] " + time.Now().String() + " " + err + " \n")
	c.JSON(status, gin.H{
		"message": message,
	})
}

func successAndExit(c *gin.Context, status int, message string) {
	os.Stdout.WriteString("[Deploy Agent Success] " + time.Now().String() + " " + message + " \n")
	c.JSON(status, gin.H{
		"message": message,
	})
}

// Get Deployment
func (dh *DeploymentHandler) GetDeployment(c *gin.Context) {
	name := c.Param("name")
	deployment, err := dh.service.Get(name)

	if err != nil {
		if err == domain.ErrDeploymentNotFound {
			errorAndExit(c, err.Error(), 404, "Not found")
			return
		}
		errorAndExit(c, err.Error(), 500, "Unexpected error")
		return
	}

	c.JSON(http.StatusOK, *deployment)
}

// Create Deployment
func (dh *DeploymentHandler) CreateDeployment(c *gin.Context) {

	var containerConfig domain.ContainerConfig

	if err := c.ShouldBindJSON(&containerConfig); err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
		return
	}

	deployment, err := dh.service.Get(containerConfig.Name)

	if err != nil && err.Error() != "not_found" {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
		return
	}

	if deployment != nil {
		errorAndExit(c, "Deployment already exists", http.StatusConflict, "Deployment already exists")
		return
	}

	deployment = &domain.Deployment{
		Name:    containerConfig.Name,
		Config:  containerConfig,
		Running: make(map[string]container.ContainerCreateCreatedBody),
	}

	for i := 0; i < deployment.Config.Replicas; i++ {

		// prepare container config
		containerProps := domain.ContainerProps{
			Image:    deployment.Config.Image,
			Name:     deployment.Config.Name + "-" + strconv.Itoa(i+1),
			Port:     deployment.Config.ContainerNet.Port + "/" + deployment.Config.ContainerNet.Proto,
			HostIP:   deployment.Config.HostNet.IP,
			HostPort: strconv.Itoa(deployment.Config.HostNet.PortFirst+i) + "/" + deployment.Config.HostNet.Proto,
			Command:  deployment.Config.Command,
			Label:    map[string]string{"by": "deploy-agent"},
		}
		containerCreateBody, err := dh.service.RunContainer(containerProps)
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "error saving deploy, please check logs")
			return
		}

		deployment.Running[containerCreateBody.ID] = containerCreateBody
	}

	if err := dh.service.Save(*deployment); err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error saving deploy, please check logs")
		return
	}

	successAndExit(c, http.StatusCreated, "deployment "+deployment.Name+" successfuly created")

}

// Stop Deployment
func (dh *DeploymentHandler) StopDeployment(c *gin.Context) {
	stopReq := struct {
		Name string `json:"name"`
	}{}

	if err := c.ShouldBindJSON(&stopReq); err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
		return
	}

	deployment, err := dh.service.Get(stopReq.Name)

	if err != nil {
		errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
		return
	}

	if deployment.Name == "" {
		c.JSON(http.StatusNotFound, stopReq.Name)
		errorAndExit(c, "deployment not found", http.StatusNotFound, "deployment not found")
		return
	}

	for k := range deployment.Running {
		err := dh.service.StopContainer(k, 60*time.Second)
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
			return
		}

		err = dh.service.DeleteContainer(k)
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
			return
		}
		delete(deployment.Running, k)

		if err := dh.service.Save(*deployment); err != nil {
			errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
			return
		}
	}

	if err := dh.service.Delete(deployment.Name); err != nil {
		errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
		return
	}

	successAndExit(c, http.StatusOK, "deployment successfully stopped")
}

func (dh *DeploymentHandler) UpdateDeployment(c *gin.Context) {
	updateParams := struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	}{}

	if err := c.ShouldBindJSON(&updateParams); err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
		return
	}

	deployment, err := dh.service.Get(updateParams.Name)
	if err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error getting deployment, please check logs")
		return
	}

	containers := make([]types.Container, 0)

	for id := range deployment.Running {
		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
		}
		container, err := dh.service.GetContainer(id)

		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
		}

		containers = append(containers, container)
	}

	for _, container := range containers {
		err := dh.service.StopContainer(container.ID, 60*time.Second)

		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
			return
		}

		err = dh.service.DeleteContainer(container.ID)

		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
			return
		}

		containerProps := domain.ContainerProps{
			Image:    updateParams.Image,
			Name:     container.Names[0],
			Port:     fmt.Sprint(container.Ports[0].PrivatePort) + "/" + deployment.Config.ContainerNet.Proto,
			HostIP:   container.Ports[0].IP,
			HostPort: strconv.Itoa(int(container.Ports[0].PublicPort)) + "/" + deployment.Config.HostNet.Proto,
			Command:  deployment.Config.Command,
			Label:    map[string]string{"by": "deploy-agent"},
		}

		createBody, err := dh.service.RunContainer(containerProps)

		if err != nil {
			errorAndExit(c, err.Error(), http.StatusInternalServerError, "Unexpected error")
			return
		}

		delete(deployment.Running, container.ID)
		deployment.Running[createBody.ID] = createBody

		dh.service.Save(*deployment)
	}

	successAndExit(c, 200, deployment.Name)

}
