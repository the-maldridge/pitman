package main

import (
	"os"

	"github.com/hashicorp/go-hclog"

	"github.com/the-maldridge/pitman/pkg/http"
)

func main() {
	appLogger := hclog.New(&hclog.LoggerOptions{
		Name:  "webui",
		Level: hclog.LevelFromString("TRACE"),
	})

	srv, err := http.New(appLogger)
	if err != nil {
		appLogger.Error("Error during webserver init", "error", err)
		os.Exit(1)
	}

	srv.Serve(":1323")
}
