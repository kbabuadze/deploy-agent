package svcs

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/kbabuadze/deploy-agent/domain"
)

type DeploymentService struct {
	repo    *domain.DeploymentsRepositoryDB
	runtime *domain.DeploymentsRuntimeDocker
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

func (ds *DeploymentService) Delete(name string) error {
	return ds.repo.Delete(name)
}

func (ds *DeploymentService) RunContainer(c domain.ContainerProps) (container.ContainerCreateCreatedBody, error) {
	return ds.runtime.RunContainer(c)
}

func (ds *DeploymentService) StopContainer(id string, timeout time.Duration) error {
	return ds.runtime.Stop(id, timeout)
}

func NewDeploymentService(repo *domain.DeploymentsRepositoryDB, runtime *domain.DeploymentsRuntimeDocker) DeploymentService {
	return DeploymentService{repo: repo, runtime: runtime}
}
