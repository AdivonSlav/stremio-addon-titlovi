package main

import (
	"go-titlovi/logger"

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

	router := buildRouter()

	err = serve(router)
	if err != nil {
		logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err.Error())
	}
}
