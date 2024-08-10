package main

import (
	"go-titlovi/api"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/titlovi"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/ristretto"
)

func main() {
	logger.InitLoggers()
	logger.LogInfo.Printf("main: initializing...")

	config.InitConfig()

	titloviClient := titlovi.NewClient(config.TitloviClientRetryAttempts, config.TitloviClientRetryDelay)

	cacheManager, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 28,
		BufferItems: 64,
	})
	if err != nil {
		logger.LogFatal.Fatalf("main: failed to initialize cache: %s", err)
	}

	router := api.BuildRouter(titloviClient, cacheManager)

	go func() {
		err = api.Serve(&router)
		if err != nil {
			logger.LogFatal.Fatalf("main: fatal error when trying to serve: %s", err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)

	<-exit
	logger.LogInfo.Printf("main: terminating...")
}
