package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ContainerProps struct {
	Image    string
	Port     string   // Container Port
	Name     string   // Container Basename
	HostIP   string   // IP to bind port on
	HostPort string   // Host Port
	Command  []string // Command that runs on container start
	Label    map[string]string
}

var ctx = context.Background()

func DeployContainer(props ContainerProps) (container.ContainerCreateCreatedBody, error) {

	containerBody := container.ContainerCreateCreatedBody{}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return containerBody, err
	}

	authBytes, _ := json.Marshal(map[string]string{
		"username": os.Getenv("DOCKER_USERNAME"),
		"password": os.Getenv("DOCKER_TOKEN"),
	})

	// Pull Container Image
	reader, err := cli.ImagePull(ctx, props.Image, types.ImagePullOptions{
		RegistryAuth: base64.StdEncoding.EncodeToString(authBytes),
	})
	if err != nil {
		return containerBody, err
	}
	io.Copy(os.Stdout, reader)

	// Host config
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(props.Port): []nat.PortBinding{
				{
					HostIP:   props.HostIP,
					HostPort: props.HostPort,
				},
			},
		},
	}

	// Container config
	containerConfig := &container.Config{
		Image:  props.Image,
		Cmd:    props.Command,
		Labels: props.Label,
	}

	// Create Container
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, props.Name)
	if err != nil {
		return containerBody, err
	}

	// Start Container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return containerBody, err
	}

	return resp, err
}

func StopContainer(id string, timeout time.Duration) error {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	if err := cli.ContainerStop(ctx, id, &timeout); err != nil {
		return err
	}

	return nil
}

func RemoveContainer(id string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if err := cli.ContainerRemove(ctx, id, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

// Get status of all containers with label "by=deploy-agent"
func GetContainers() ([]types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	filter := filters.NewArgs()

	filter.Add("label", "by=deploy-agent")

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filter})

	return containers, err
}

func GetContainer(id string) (types.Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	filter := filters.NewArgs()

	filter.Add("id", id)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filter})

	return containers[0], err

}
