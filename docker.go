package main

import (
	"context"
	"fmt"
	"io"
	"log"
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

	// Pull Container Image
	reader, err := cli.ImagePull(ctx, props.Image, types.ImagePullOptions{})
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

	return resp, nil
}

func UpdateContainers(name string, image string, updateSlice []types.Container) error {

	for _, container := range updateSlice {

		if err := StopContainer(container.ID, 60*time.Second); err != nil {
			log.Println(err.Error())
			return err
		}

		if err := RemoveContainer(container.ID); err != nil {
			log.Println(err.Error())
			return err
		}
		delete(containers, container.ID)

		props := ContainerProps{
			Image:    image,
			Name:     container.Names[0],
			Port:     fmt.Sprint(container.Ports[0].PrivatePort) + "/tcp",
			HostIP:   fmt.Sprint(container.Ports[0].IP),
			HostPort: fmt.Sprint(container.Ports[0].PublicPort) + "/tcp",
			Command:  []string{"nginx", "-g", "daemon off;"},
			Label:    map[string]string{"by": "deploy-agent"},
		}

		newContainer, err := DeployContainer(props)
		containers[newContainer.ID] = newContainer
		if err != nil {
			log.Println(err.Error())
			return err
		}

	}

	return nil
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
