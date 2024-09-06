package helper

import (
	"fmt"
	"log/slog"
	"sync"
)

type AppWorker struct {
	env    string
	logger *slog.Logger
	Wg     *sync.WaitGroup
}

func NewAppWorker(logger *slog.Logger, env string, wg *sync.WaitGroup) *AppWorker {
	return &AppWorker{
		env:    env,
		logger: logger,
		Wg:     wg,
	}
}

// The background() helper accepts an arbitrary function as a parameter.
func (app *AppWorker) Background(fn func()) {
	// Increment the WaitGroup counter.
	app.Wg.Add(1)

	// Launch a background goroutine.
	go func() {
		// Use defer to decrement the WaitGroup counter before the goroutine returns.
		defer app.Wg.Done()

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
