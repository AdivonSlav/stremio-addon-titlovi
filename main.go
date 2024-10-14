package main

import (
	"context"
	"errors"
	"go-titlovi/api"
	"go-titlovi/internal/config"
	"go-titlovi/internal/logger"
	"go-titlovi/internal/titlovi"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dgraph-io/ristretto"
)

var (
	Build = ""
)

func main() {
	logger.InitLoggers()
	logger.LogInfo.Printf("main: initializing...")
	logger.LogInfo.Printf("main: build %s", Build)

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
	server := api.BuildServer(&router)

	go func() {
		logger.LogInfo.Printf("main: listening on port %s", config.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.LogFatal.Fatalf("main: error when trying to serve: %s", err)
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGTERM, syscall.SIGINT)

	<-exit
	logger.LogInfo.Printf("main: terminating...")

	ctx, release := context.WithTimeout(context.Background(), 10*time.Second)
	defer release()

	if err := server.Shutdown(ctx); err != nil {
		logger.LogFatal.Fatalf("main: error when trying to shutdown server: %s", err)
	}
	logger.LogInfo.Printf("main: terminated")
}
