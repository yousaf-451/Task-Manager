// Command api is the entrypoint for the Task Management REST API.
//
// Request flow:
//
//	React (frontend) --HTTP--> routes --> handler --> service --> repository --> MySQL
//
// Each arrow is a strict boundary: handlers never touch SQL, and the
// repository never knows about HTTP. This keeps the codebase easy to test
// and to extend (e.g. swapping MySQL for Postgres would only touch the
// repository package).
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/granet/task-manager/internal/config"
	"github.com/granet/task-manager/internal/handler"
	"github.com/granet/task-manager/internal/repository"
	"github.com/granet/task-manager/internal/routes"
	"github.com/granet/task-manager/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	// 1. Configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 2. Database connection pool
	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	// 3. Wire layers: repository -> service -> handler -> router
	taskRepo := repository.NewMySQLTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	userRepo := repository.NewMySQLUserRepository(db)
	sessionRepo := repository.NewMySQLSessionRepository(db)
	authService := service.NewAuthService(userRepo, sessionRepo)
	authHandler := handler.NewAuthHandler(authService, cfg.CookieSecure)

	router := routes.NewRouter(taskHandler, authHandler, authService, cfg.CORSAllowedOrigins)

	// 4. HTTP server with sane timeouts
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start server in a goroutine so we can listen for shutdown signals.
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("task-manager API listening on %s", cfg.Addr())
		serverErrors <- srv.ListenAndServe()
	}()

	// 6. Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return err
		}
	case sig := <-quit:
		log.Printf("received signal %s, shutting down gracefully...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		log.Println("server stopped cleanly")
	}

	return nil
}

// openDB opens the MySQL connection pool and verifies connectivity with a
// ping, using a short retry loop so `docker compose up` (where MySQL may
// still be starting) doesn't crash the API on first boot.
func openDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetimeMin)

	var pingErr error
	for attempt := 1; attempt <= 10; attempt++ {
		pingErr = db.Ping()
		if pingErr == nil {
			log.Println("connected to MySQL")
			return db, nil
		}
		log.Printf("waiting for MySQL (attempt %d/10): %v", attempt, pingErr)
		time.Sleep(2 * time.Second)
	}

	return nil, pingErr
}
