package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

// Recieves create signal, Unmarshals JSON, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleCreate(db *bolt.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		var containerConfig ContainerConfig
		var containerBody container.ContainerCreateCreatedBody
		var err error
		var deployment = Deployment{}

		if err := c.BindJSON(&containerConfig); err != nil {
			panic(err)
		}

		deployment.get(db, containerConfig.Name)

		fmt.Printf("deployment --- %v", deployment)

		if deployment.Name != "" {
			fmt.Println("deployment already exists")
			return
		}

		deployment = Deployment{
			Name:    containerConfig.Name,
			Config:  containerConfig,
			Running: make(map[string]container.ContainerCreateCreatedBody),
		}

		deployment.save(db)

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

				deployment.Running[containerBody.ID] = containerBody

				deployment.save(db)

				containers[containerBody.ID] = containerBody
			}
		}()

		c.JSON(http.StatusOK, containerConfig)
	}
}

type StopRequest struct {
	Name string `json:"name"`
}

func handleStopDeploy(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		stopReq := StopRequest{}

		if err := c.BindJSON(&stopReq); err != nil {
			panic(err)
		}

		name := []byte("")
		db.View(func(tx *bolt.Tx) error {
			// Assume bucket exists and has keys
			b := tx.Bucket([]byte("Deployments"))

			name = b.Get([]byte(stopReq.Name))

			return nil
		})

		if name == nil {
			c.JSON(http.StatusNotFound, name)
			fmt.Printf("No such deployment")
			return
		}

		deployment := Deployment{}

		err := json.Unmarshal([]byte(name), &deployment)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for k := range deployment.Running {
			fmt.Println("Stopping" + k)
			err = StopContainer(k, 60*time.Second)
			if err != nil {
				fmt.Println(err.Error())
			}

			err = RemoveContainer(k)
			if err != nil {
				fmt.Println(err.Error())
			}
			delete(deployment.Running, k)

			db.Update(func(tx *bolt.Tx) error {
				b, err := tx.CreateBucketIfNotExists([]byte("Deployments"))

				if err != nil {
					return fmt.Errorf("create bucket: %s", err)
				}

				encoded, err := json.Marshal(deployment)
				if err != nil {
					return err
				}

				err = b.Put([]byte(deployment.Name), encoded)

				if err != nil {
					return fmt.Errorf("create bucket: %s", err)
				}

				return nil

			})

		}

		if len(deployment.Running) == 0 {
			fmt.Println("Deleting Deployment")
			deployment.delete(db)
		}

		c.JSON(http.StatusOK, stopReq)
	}
}

func handleReset(db *bolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		db.Update(func(t *bolt.Tx) error {
			b := t.Bucket([]byte("Deployemnets"))
			b.Delete([]byte("nginx"))
			return nil
		})

		c.JSON(http.StatusOK, gin.H{
			"status": "reset",
		})
	}
}

// Recieves stop signal, passes action to goroutine and immediately returns a result.
// Progress can be obtained though /status endpoint
func handleStop(c *gin.Context) {

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
