package main

import (
	"log/slog"
	"os"

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

	// Declare an instance of the config struct.
	var cfg config.Config
	err := cfg.InitConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
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

	// Declare an instance of the application struct, containing the config struct and
	// the logger.
	app := &config.Application{
		Config: cfg,
		Logger: logger,
		Errors: helper.NewAppErrors(logger, cfg.Env.String()),
		Models: data.NewModels(db),
		Mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
	}

	// Call app.serve() to start the server.
	err = server.Serve(app)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
