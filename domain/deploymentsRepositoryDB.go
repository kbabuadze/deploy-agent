package domain

import (
	"encoding/json"
	"errors"

	bolt "go.etcd.io/bbolt"
)

type DeploymentsRepositoryDB struct {
	client *bolt.DB
}

func (dr *DeploymentsRepositoryDB) Save(d Deployment) error {
	return dr.client.Update(func(tx *bolt.Tx) error {
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

func (dr *DeploymentsRepositoryDB) Get(name string) (*Deployment, error) {

	deployment := Deployment{}

	err := dr.client.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Deployments"))

		result := b.Get([]byte(name))

		if len(result) == 0 {
			return errors.New("not_found")
		}

		err := json.Unmarshal(result, &deployment)
		if err != nil {
			return err
		}

		return nil
	})

	return &deployment, err
}

func NewDeploymentsRepositoryDB(client *bolt.DB) DeploymentsRepositoryDB {
	return DeploymentsRepositoryDB{client: client}
}
