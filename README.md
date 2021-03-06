# Deploy Agent

> **Disclaimer**: This project is ***NOT*** intended for a production evironment. Use at your own risk. 

Depoy agent can run on a remote machine to pull and run containers. 

### Prerequisites

You need to have the latest versions of [Go](https://golang.org/doc/install) and [Docker](https://docs.docker.com/engine/install/)

### Installation

Build an executable:
```go
go build . 

```
Specify IP and PORT in .env  using the same format as in the .env.example 
or:
```bash
cp .env.example .env
```
Run an agent: 
```bash
./deploy-agent
```
### Usage:



#### Create a deployment from a JSON object: 
`POST /create` 
```curl

curl -XPOST http://localhost:8008/create -d '{
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
  "replicas": 2,
  "command": [
    "nginx",
    "-g",
    "daemon off;"
  ]
}' -u admin:admin

```

#### Get running containers:
`GET /get`
````curl
 curl -XGET http://localhost:8008/get/nginx -u admin:admin
````

#### Update deployment image:
 `PATCH /update`
```curl
curl -XPATCH localhost:8008/update -d '{"name":"nginx","image":"nginx:1.21.0"}' -u admin:admin
```

#### Stop deployment:
`POST /stop`
```curl
 curl -XPOST http://localhost:8008/stop -d '{"name":"nginx"}' -u admin:admin
 ```
