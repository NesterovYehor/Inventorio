package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/database"
	"github.com/NesterovYehor/Inventorio/internal/handlers"
	"github.com/NesterovYehor/Inventorio/ui"
)

type App struct {
	db *database.DB
	s  *http.Server
}

func Setup(dbPath string, port string) (*App, error) {
	db, err := database.Connect(dbPath)
	if err != nil {
		return nil, fmt.Errorf("Error during setup: %w", err)
	}

	ui.InitUI()

	h := handlers.New(db)
	server := &http.Server{
		Addr:    port,
		Handler: h.Routes(),
	}
	return &App{db: db, s: server}, nil
}

func (app *App) Run(ctx context.Context) error {
	svrErr := make(chan error, 1)
	log.Println("Server started")

	go func() {
		if err := app.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			svrErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("\nShutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		defer cancel()

		if err := app.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("Failed to gracefully stop application: %w", err)
		}

	case err := <-svrErr:
		return fmt.Errorf("server error:%w", err)
	}
	return nil

}

func (app *App) Stop(ctx context.Context) error {
	var shutdownErr error
	if err := app.s.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		app.s.Close()
		shutdownErr = fmt.Errorf("could not stop server gracefully: %w", err)
	}

	if err := app.db.Close(); err != nil {
		if shutdownErr != nil {
			return fmt.Errorf("%v; additionally, failed to close SQLite connection: %w", shutdownErr, err)
		}
		return fmt.Errorf("could not close connecion to Sqlite greacefully: %w", err)
	}
	return nil

}
