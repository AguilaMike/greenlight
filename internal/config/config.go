package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"

	"github.com/AguilaMike/greenlight/internal/data"
	"github.com/AguilaMike/greenlight/internal/mailer"
	"github.com/AguilaMike/greenlight/internal/vcs"
	"github.com/AguilaMike/greenlight/pkg/utilities/rest/helper"
)

// Declare a string containing the application version number. Later in the book we'll
// generate this automatically at build time, but for now we'll just store the version
// number as a hard-coded global constant.
const (
	API_VERSION = "v1"
)

var VERSION = vcs.Version()

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
	Port int     `env:"PORT" flag:"port" default:"4000" desc:"API server port"`
	Env  EnvType `env:"ENV" flag:"env" default:"development" desc:"Environment (development|staging|production)"`
	Db   struct {
		Dsn          string        `env:"DB_DSN" flag:"db-dsn" default:"postgres://greenlight:pa55word@localhost:5433/greenlight?sslmode=disable" desc:"PostgreSQL DSN"`
		MaxOpenConns int           `env:"DB_MAX_OPEN_CONNS" flag:"db-max-open-conns" default:"25" desc:"PostgreSQL max open connections"`
		MaxIdleConns int           `env:"DB_MAX_IDLE_CONNS" flag:"db-max-idle-conns" default:"25" desc:"PostgreSQL max idle connections"`
		MaxIdleTime  time.Duration `env:"DB_MAX_IDLE_TIME" flag:"db-max-idle-time" default:"15m" desc:"PostgreSQL max connection idle time"`
	}
	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	Limiter struct {
		Rps     float64 `env:"LIMITER_RPS" flag:"limiter-rps" default:"2" desc:"Rate limiter maximum requests per second"`
		Burst   int     `env:"LIMITER_BURST" flag:"limiter-burst" default:"4" desc:"Rate limiter maximum burst"`
		Enabled bool    `env:"LIMITER_ENABLED" flag:"limiter-enabled" default:"true" desc:"Enable rate limiter"`
	}
	Smtp struct {
		Host     string `env:"SMTP_HOST" flag:"smtp-host" default:"sandbox.smtp.mailtrap.io" desc:"SMTP host"`
		Port     int    `env:"SMTP_PORT" flag:"smtp-port" default:"25" desc:"SMTP port"`
		Username string `env:"SMTP_USERNAME" flag:"smtp-username" default:"a7420fc0883489" desc:"SMTP username"`
		Password string `env:"SMTP_PASSWORD" flag:"smtp-password" default:"e75ffd0a3aa5ec" desc:"SMTP password"`
		Sender   string `env:"SMTP_SENDER" flag:"smtp-sender" default:"Greenlight <no-reply@greenlight.net>" desc:"SMTP sender"`
	}
	// Add a cors struct and trustedOrigins field with the type []string.
	Cors struct {
		TrustedOrigins []string `env:"CORS_TRUSTED_ORIGINS" flag:"cors-trusted-origins" default:"http://localhost:4000" desc:"CORS trusted origins"`
	}
}

func (c *Config) InitConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	err = loadStructConfig(c, c)
	if err != nil {
		return err
	}

	flag.Parse()

	return nil
}

func loadStructConfig(cfg interface{}, c *Config) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct {
			// Recursively initialize nested structs
			nestedCfg := field.Addr().Interface()
			if err := loadStructConfig(nestedCfg, c); err != nil {
				return err
			}
			continue
		}

		envTag := fieldType.Tag.Get("env")
		flagTag := fieldType.Tag.Get("flag")
		defaultTag := fieldType.Tag.Get("default")
		descTag := fieldType.Tag.Get("desc")

		if envTag != "" {
			envValue := os.Getenv(envTag)
			if envValue != "" {
				if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
					elemType := field.Type().Elem().Kind()
					if elemType == reflect.String {
						elemValues := []string{}
						for _, value := range strings.Split(envValue, " ") {
							if strings.TrimSpace(value) != "" {
								elemValues = append(elemValues, value)
							}
						}
						field.Set(reflect.ValueOf(elemValues))
					} else {
						return errors.New(fmt.Sprintf("invalid type: %s", elemType))
					}
					continue
				}

				switch field.Kind() {
				case reflect.String:
					if envTag != "ENV" {
						field.SetString(envValue)
					} else {
						err := c.SetEnv(EnvType(envValue))
						if err != nil {
							return err
						}
					}
				case reflect.Int:
					intValue, err := strconv.Atoi(envValue)
					if err != nil {
						return err
					}
					field.SetInt(int64(intValue))
				case reflect.Int64:
					intValue, err := strconv.ParseInt(envValue, 10, 64)
					if err != nil {
						return err
					}
					field.SetInt(int64(intValue))
				case reflect.Float64:
					int64Value, err := strconv.ParseFloat(envValue, 64)
					if err != nil {
						return err
					}
					field.SetFloat(int64Value)
				case reflect.Bool:
					boolValue, err := strconv.ParseBool(envValue)
					if err != nil {
						return err
					}
					field.SetBool(boolValue)
				default:
					return errors.New(fmt.Sprintf("invalid type: %s", field.Kind()))
				}
			} else if flagTag != "" {
				if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
					elemType := field.Type().Elem().Kind()
					if elemType == reflect.String {
						elemValues := []string{}
						for _, value := range strings.Split(defaultTag, " ") {
							if strings.TrimSpace(value) != "" {
								elemValues = append(elemValues, value)
							}
						}
						field.Set(reflect.ValueOf(elemValues))
					} else {
						return errors.New(fmt.Sprintf("invalid type: %s", elemType))
					}
					continue
				}

				switch field.Kind() {
				case reflect.String:
					if envTag != "ENV" {
						flag.StringVar(field.Addr().Interface().(*string), flagTag, defaultTag, descTag)
					} else {
						err := c.SetEnv(EnvType(defaultTag))
						if err != nil {
							return err
						}
					}
				case reflect.Int:
					intDefault, _ := strconv.Atoi(defaultTag)
					flag.IntVar(field.Addr().Interface().(*int), flagTag, intDefault, descTag)
				case reflect.Float64:
					floatDefault, _ := strconv.ParseFloat(defaultTag, 64)
					flag.Float64Var(field.Addr().Interface().(*float64), flagTag, floatDefault, descTag)
				case reflect.Bool:
					boolDefault, _ := strconv.ParseBool(defaultTag)
					flag.BoolVar(field.Addr().Interface().(*bool), flagTag, boolDefault, descTag)
				case reflect.Int64:
					if field.Type().String() == "time.Duration" {
						durationDefault, _ := time.ParseDuration(defaultTag)
						flag.DurationVar(field.Addr().Interface().(*time.Duration), flagTag, durationDefault, descTag)
					} else {
						intDefault, _ := strconv.ParseInt(defaultTag, 10, 64)
						flag.Int64Var(field.Addr().Interface().(*int64), flagTag, intDefault, descTag)
					}
				default:
					return errors.New(fmt.Sprintf("invalid type: %s", field.Kind()))
				}
			}
		}
	}

	return nil
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

// String returns the environment as a string
func (e EnvType) String() string {
	return string(e)
}

// Define an application struct to hold the dependencies for our HTTP handlers, helpers,
// and middleware. At the moment this only contains a copy of the config struct and a
// logger, but it will grow to include a lot more as our build progresses.
type Application struct {
	Config Config
	Logger *slog.Logger
	Errors *helper.AppErrors
	Worker *helper.AppWorker
	Models data.Models
	Mailer mailer.Mailer
	Wg     *sync.WaitGroup
}
