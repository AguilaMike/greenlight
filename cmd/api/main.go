package main

import (
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/AguilaMike/greenlight/internal/config"
	"github.com/AguilaMike/greenlight/internal/data"
	"github.com/AguilaMike/greenlight/internal/database"
	"github.com/AguilaMike/greenlight/internal/mailer"
	"github.com/AguilaMike/greenlight/internal/server"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

func main() {
	// Initialize a new structured logger which writes log entries to the standard out
	// stream.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create a new version boolean flag with the default value of false.
	displayVersion := flag.Bool("version", false, "Display version and exit")

	// Declare an instance of the config struct.
	var cfg config.Config
	err := cfg.InitConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", config.VERSION)
		os.Exit(0)
	}

	// Call the openDB() helper function (see below) to create the connection pool,
	// passing in the config struct. If this returns an error, we log it and exit the
	// application immediately.
	db, err := database.OpenDB(&cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the
	// main() function exits.
	defer db.Close()

	// Also log a message to say that the connection pool has been successfully
	// established.
	logger.Info("database connection pool established")

	migrationDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://scripts/migrations", "postgres", migrationDriver)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	err = migrator.Up()
	if err != nil && err != migrate.ErrNoChange {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database migrations applied")
	wg := &sync.WaitGroup{}

	// Publish a new "version" variable in the expvar handler containing our application
	// version number (currently the constant "1.0.0").
	expvar.NewString("version").Set(config.VERSION)

	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &config.Application{
		Config: cfg,
		Logger: logger,
		Errors: helper.NewAppErrors(logger, cfg.Env.String()),
		Worker: helper.NewAppWorker(logger, cfg.Env.String(), wg),
		Models: data.NewModels(db),
		Mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
		Wg:     wg,
	}

	// Call app.serve() to start the server.
	err = server.Serve(app)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
