package main

import (
	"go-titlovi/api"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/titlovi"
	"os"
	"os/signal"
	"syscall"

	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func main() {
	logger.InitLoggers()
	logger.LogInfo.Printf("main: initializing...")

	config.InitConfig()

	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(config.MemoryCacheAlgorithm),
		memory.AdapterWithCapacity(config.MemoryCacheMaxResponses),
	)
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to initialize memory cache adapter: %s", err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(config.MemoryCacheTTL),
		cache.ClientWithRefreshKey(config.MemoryCacheRefreshKey),
	)
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to initialize cache: %s", err)
	}

	titloviClient := titlovi.NewClient(config.TitloviClientRetryAttempts, config.TitloviClientRetryDelay)

	router := api.BuildRouter(titloviClient, cacheClient)

	go func() {
		err = api.Serve(&router)
		if err != nil {
			logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err.Error())
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)

	<-exit
	logger.LogInfo.Printf("Terminating...")
}
