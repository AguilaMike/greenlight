package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/rest/routes"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

func main() {
	// Declare an instance of the config struct.
	var cfg config.Config
	var enviroment string

	// Initialize a new structured logger which writes log entries to the standard out
	// stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Read the value of the port and env command-line flags into the config struct. We
	// default to using the port number 4000 and the environment "development" if no
	// corresponding flags are provided.
	flag.IntVar(&cfg.Port, "port", 4000, "API server port")
	flag.StringVar(&enviroment, "env", "development", "Environment (development|staging|production)")
	flag.Parse()
	if err := cfg.SetEnv(config.EnvType(enviroment)); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &config.Application{
		Config: cfg,
		Logger: logger,
		Errors: helper.NewAppErrors(logger, cfg.Env.String()),
	}

	// Declare a HTTP server which listens on the port provided in the config struct,
	// uses the servemux we created above as the handler, has some sensible timeout
	// settings and writes any log messages to the structured logger at Error level.
	// Use the httprouter instance returned by app.routes() as the server handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      routes.GenerateRoutes(app),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Start the HTTP server.
	logger.Info("starting server", "addr", srv.Addr, "env", cfg.Env)

	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
