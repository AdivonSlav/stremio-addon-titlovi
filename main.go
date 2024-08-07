package main

import (
	"go-titlovi/logger"
	"go-titlovi/titlovi"

	"github.com/joho/godotenv"
)

func main() {
	logger.InitLoggers()
	logger.LogInfo.Printf("main: initializing...")

	err := godotenv.Load()
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to load environment file\n")
	}
	initConfig()

	client := titlovi.NewClient(TitloviUsername, TitloviPassword)

	router := buildRouter(client)

	err = serve(router)
	if err != nil {
		logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err.Error())
	}
}
