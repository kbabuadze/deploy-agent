package domain

import (
	"github.com/docker/docker/api/types/container"
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

type ContainerConfig struct {
	Image string `json:"image"`
	Name  string `json:"name"`

	ContainerNet struct {
		Port  string `json:"port"`
		Proto string `json:"proto"`
	} `json:"containerNet"`

	HostNet struct {
		IP        string `json:"ip"`
		PortFirst int    `json:"portFirst"`
		Proto     string `json:"proto"`
	} `json:"hostNet"`

	Replicas int      `json:"replicas"`
	Command  []string `json:"command"`
}

type Deployment struct {
	Name    string                                          `json:"name"`
	Config  ContainerConfig                                 `json:"config"`
	Running map[string]container.ContainerCreateCreatedBody `json:"running"`
}
