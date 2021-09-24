package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kbabuadze/deploy-agent/svcs"
)

type DeploymentHandler struct {
	service svcs.DeploymentService
}

func (dh *DeploymentHandler) GetDeployments(c *gin.Context) {
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
