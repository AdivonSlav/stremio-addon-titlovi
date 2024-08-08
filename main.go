package main

import (
	"go-titlovi/common"
	"go-titlovi/logger"
	"go-titlovi/titlovi"
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
	common.InitConfig()

	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to initialize cache: %s", err)
	}
	client := titlovi.NewClient(common.TitloviUsername, common.TitloviPassword)

	router := common.BuildRouter(client, cache)

	err = common.Serve(router)
	if err != nil {
		logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err.Error())
	}
}
