package domain

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DeploymentsRuntimeDocker struct {
	ctx *context.Context
}

// func (dr *DeploymentsRuntimeDocker) Run(d *Deployment) error {
// 	for i := 0; i < d.Config.Replicas; i++ {

// 		// prepare container config
// 		containerProps := ContainerProps{
// 			Image:    d.Config.Image,
// 			Name:     d.Config.Name + "-" + strconv.Itoa(i+1),
// 			Port:     d.Config.ContainerNet.Port + "/" + d.Config.ContainerNet.Proto,
// 			HostIP:   d.Config.HostNet.IP,
// 			HostPort: strconv.Itoa(d.Config.HostNet.PortFirst+i) + "/" + d.Config.HostNet.Proto,
// 			Command:  d.Config.Command,
// 			Label:    map[string]string{"by": "deploy-agent"},
// 		}

// 		containerCreateBody, err := dr.RunContainer(containerProps)

// 		if err != nil {
// 			return err
// 		}

// 		d.Running[containerCreateBody.ID] = containerCreateBody

// 	}

// 	return nil
// }

func (dr *DeploymentsRuntimeDocker) RunContainer(props ContainerProps) (container.ContainerCreateCreatedBody, error) {
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
	reader, err := cli.ImagePull(*dr.ctx, props.Image, types.ImagePullOptions{
		RegistryAuth: base64.StdEncoding.EncodeToString(authBytes),
	})
	if err != nil {
		return containerBody, err
	}
	io.Copy(os.Stdout, reader)
	fmt.Println("")

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
	resp, err := cli.ContainerCreate(*dr.ctx, containerConfig, hostConfig, nil, nil, props.Name)
	if err != nil {
		return containerBody, err
	}

	// Start Container
	if err := cli.ContainerStart(*dr.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return containerBody, err
	}

	return resp, err
}

func NewDeploymentsRuntime(ctx *context.Context) DeploymentsRuntimeDocker {
	return DeploymentsRuntimeDocker{ctx: ctx}
}
