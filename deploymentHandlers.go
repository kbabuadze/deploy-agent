package main

import (
	"net/http"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"github.com/kbabuadze/deploy-agent/domain"
	"github.com/kbabuadze/deploy-agent/svcs"
)

type DeploymentHandler struct {
	service svcs.DeploymentService
}

func (dh *DeploymentHandler) GetDeployment(c *gin.Context) {
	name := c.Param("name")
	deployment, err := dh.service.Get(name)

	if err != nil {
		if err.Error() == "not_found" {
			errorAndExit(c, err.Error(), 404, "Not found")
			return
		}
		errorAndExit(c, err.Error(), 500, "Unexpected error")
		return
	}

	c.JSON(http.StatusOK, *deployment)
}

func (dh *DeploymentHandler) CreateDeployment(c *gin.Context) {

	var containerConfig domain.ContainerConfig
	var deployment *domain.Deployment

	if err := c.ShouldBindJSON(&containerConfig); err != nil {
		errorAndExit(c, err.Error(), http.StatusInternalServerError, "error decoding config, please check logs")
		return
	}

	_, err := dh.service.Get(containerConfig.Name)

	if err != nil && err.Error() != "not_found" {
		errorAndExit(c, err.Error(), http.StatusNotFound, "Unexpected error")
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
