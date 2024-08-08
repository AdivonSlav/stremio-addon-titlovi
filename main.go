package main

import (
	"go-titlovi/api"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/titlovi"
	"time"

	"github.com/allegro/bigcache"
	"github.com/joho/godotenv"
)

func main() {
	logger.InitLoggers()
	logger.LogInfo.Printf("main: initializing...")

	err := godotenv.Load()
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to load environment file\n")
	}
	config.InitConfig()

	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to initialize cache: %s", err)
	}
	client := titlovi.NewClient(config.TitloviUsername, config.TitloviPassword)

	router := api.BuildRouter(client, cache)

	err = api.Serve(router)
	if err != nil {
		logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err.Error())
	}
}
