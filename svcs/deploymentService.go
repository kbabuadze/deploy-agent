package svcs

import (
	"github.com/kbabuadze/deploy-agent/domain"
)

type DeploymentService struct {
	repo *domain.DeploymentsRepositoryDB
}

func (ds *DeploymentService) Save(d domain.Deployment) error {
	return ds.repo.Save(d)
}

func (ds *DeploymentService) Get(name string) (*domain.Deployment, error) {
	deployment, err := ds.repo.Get(name)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func NewDploymentService(repo *domain.DeploymentsRepositoryDB) DeploymentService {
	return DeploymentService{repo: repo}
}
