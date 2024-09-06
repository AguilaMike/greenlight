package helper

import (
	"fmt"
	"log/slog"
)

type AppWorker struct {
	env    string
	logger *slog.Logger
}

func NewAppWorker(logger *slog.Logger, env string) *AppWorker {
	return &AppWorker{
		env:    env,
		logger: logger,
	}
}

// The background() helper accepts an arbitrary function as a parameter.
func (app *AppWorker) Background(fn func()) {
	// Launch a background goroutine.
	go func() {
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		fn()
	}()
}
