package main

import (
	"log"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("main: initializing...")

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("main: failed to load environment file\n")
	}
	initConfig()

	router := buildRouter()

	err = serve(router)
	if err != nil {
		log.Fatalf("main: fatal error when trying to serve: %s", err.Error())
	}
}
