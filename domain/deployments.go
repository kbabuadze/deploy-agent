package domain

import (
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	bolt "go.etcd.io/bbolt"
)

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

// Saves Deployment to BoltDB
func (d *Deployment) Save(db *bolt.DB) error {

	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Deployments"))

		if err != nil {
			return err
		}

		encoded, err := json.Marshal(d)
		if err != nil {
			return err
		}

		err = b.Put([]byte(d.Name), encoded)

		if err != nil {
			return err
		}

		return nil

	})
}

// Gets deployment from BoltDB
func (d *Deployment) Get(db *bolt.DB, name string) error {

	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Deployments"))

		err := json.Unmarshal(b.Get([]byte(name)), d)

		if err != nil {
			return err
		}

		return nil
	})
}
