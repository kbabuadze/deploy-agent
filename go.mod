module github.com/kbabuadze/deploy-agent

go 1.16

require (
	github.com/containerd/containerd v1.5.4 // indirect
	github.com/docker/docker v20.10.6+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gin-gonic/gin v1.7.1
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.etcd.io/bbolt v1.3.5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/grpc v1.37.1 // indirect
)


replace github.com/kbabuadze/deploy-agent/svcs => /home/ubuntu/for-fun/deploy-agent/svcs
replace github.com/kbabuadze/deploy-agent/domain => /home/ubuntu/for-fun/deploy-agent/domain
