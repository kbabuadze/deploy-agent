# Deploy Agent

> **Disclaimer**: This project is ***NOT*** intended for a production evironment. Use at your own risk. 

Depoy agent can be run on a remote machine to pull and run containers. 

### Prerequisites

You need to have latest versions of [GO](https://golang.org/doc/install) and [Docker](https://docs.docker.com/engine/install/)

### Installation

Build an executable:
```go
go build . 

```
Specify IP and PORT in .env  using the same format as in .env.example 
or:
```bash
cp .env.example .env
```
Run an agent: 
```bash
./deploy-agent
```
### Requests

Create a deploymnet from a JSON object: 
```curl

curl localhost:8008/create -d '{
  "image": "nginx:latest",
  "name": "nginx",
  "containerNet": {
    "port": "80",
    "proto": "tcp"
  },
  "hostNet": {
    "ip": "0.0.0.0",
    "portFirst": 8090,
    "proto": "tcp"
  },
  "replicas": 3,
  "command": [
    "nginx",
    "-g",
    "daemon off;"
  ]
}'

```
