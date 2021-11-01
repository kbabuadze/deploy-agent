package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kbabuadze/deploy-agent/app"
)

func main() {

	//Setup .env
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	agent := app.Agent{
		BasicAuthUser: os.Getenv("DEPLOY_USERNAME"),
		BasicAuthPass: os.Getenv("DEPLOY_PASSWORD"),
		Port:          os.Getenv("LISTEN_ON"),
		DBName:        "my.db",
	}

	app.Run(&agent)
}
