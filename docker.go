package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type ContainerProps struct {
	Image    string
	Port     string   // Container Port
	HostIP   string   // IP to bind port on
	HostPort string   // Host Port
	Command  []string // Command that runs on container start
}

func DeployContainer(props ContainerProps) (container.ContainerCreateCreatedBody, error) {

	ctx := context.Background()

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
		Image: props.Image,
		Cmd:   props.Command,
	}

	// Create Container
	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return containerBody, err
	}

	// Start Container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return containerBody, err
	}

	return resp, nil
}

func StopContainer(id string, timeout time.Duration) error {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStop(ctx, id, &timeout); err != nil {
		return err
	}

	return nil
}
