package config

import (
	"errors"
	"fmt"
	"log/slog"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const (
	VERSION     = "1.0.0"
	API_VERSION = "v1"
)

// Define a config struct to hold all the configuration settings for our application.
// For now, the only configuration settings will be the network port that we want the
// server to listen on, and the name of the current operating environment for the
// application (development, staging, production, etc.). We will read in these
// configuration settings from command-line flags when the application starts.
// EnvType is a custom type for environment
type EnvType string

const (
	Development EnvType = "development"
	Staging     EnvType = "staging"
	Production  EnvType = "production"
)

type Config struct {
	Port int
	Env  EnvType
}

// SetEnv sets the environment and validates it
func (c *Config) SetEnv(env EnvType) error {
	switch env {
	case Development, Staging, Production:
		c.Env = env
		return nil
	default:
		return errors.New(fmt.Sprintf("invalid environment: %s", env))
	}
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type Application struct {
	Config Config
	Logger *slog.Logger
}
