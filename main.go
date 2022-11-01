package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/pitman/pkg/http"
	"github.com/the-maldridge/pitman/pkg/kv"
)

func main() {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "pitman",
		Level: hclog.LevelFromString(os.Getenv("LOG_LEVEL")),
	})

	var store http.KV
	storeImpl := strings.ToLower(os.Getenv("PITMAN_STORE"))
	if storeImpl == "" {
		storeImpl = "redis"
	}
	switch storeImpl {
	case "redis":
		store = kv.NewRedis()
	}

	srv, err := http.New(http.WithLogger(appLogger), http.WithStorage(store))
	if err != nil {
		appLogger.Error("Error during webserver init", "error", err)
		os.Exit(1)
	}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig
		appLogger.Info("Interrupt received, shutting down")

		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				appLogger.Error("Graceful shutdown timed out.. forcing exit.")
				os.Exit(5)
			}
		}()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			appLogger.Error("Error occured during shutdown", "error", err)
		}
		serverStopCtx()
	}()

	bind := os.Getenv("PITMAN_ADDR")
	if bind == "" {
		bind = ":1323"
	}
	srv.Serve(bind)
	<-serverCtx.Done()
}
