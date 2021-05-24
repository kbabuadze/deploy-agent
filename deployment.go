package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

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

// Save deployemnt to bolt
func (d *Deployment) save(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("Deployments"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
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

// Get deployment from boltDB
func (d *Deployment) get(db *bolt.DB, name string) error {

	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Deployments"))

		err := json.Unmarshal(b.Get([]byte(name)), d)

		if err != nil {
			return err
		}

		return nil
	})
}

// Delete deployment from boltDB
func (d *Deployment) delete(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("Deployments"))
		b.Delete([]byte(d.Name))

		return nil
	})
}

func (d *Deployment) run(db *bolt.DB) {

	for i := 0; i < d.Config.Replicas; i++ {

		// prepare container config
		containerProps := ContainerProps{
			Image:    d.Config.Image,
			Name:     d.Config.Name + "-" + strconv.Itoa(i+1),
			Port:     d.Config.ContainerNet.Port + "/" + d.Config.ContainerNet.Proto,
			HostIP:   d.Config.HostNet.IP,
			HostPort: strconv.Itoa(d.Config.HostNet.PortFirst+i) + "/" + d.Config.HostNet.Proto,
			Command:  d.Config.Command,
			Label:    map[string]string{"by": "deploy-agent"},
		}

		// run
		containerBody, err := DeployContainer(containerProps)
		if err != nil {
			log.Println(err.Error())
			return
		}

		d.Running[containerBody.ID] = containerBody

		// save
		d.save(db)
	}

}

func (d *Deployment) stop(db *bolt.DB) {
	for k := range d.Running {
		fmt.Println("Stopping " + k)
		err := StopContainer(k, 60*time.Second)
		if err != nil {
			fmt.Println(err.Error())
		}

		err = RemoveContainer(k)
		if err != nil {
			fmt.Println(err.Error())
		}
		delete(d.Running, k)

		d.save(db)

	}

	if len(d.Running) == 0 {
		fmt.Println("Deleting Deployments")
		d.delete(db)
	}
}
