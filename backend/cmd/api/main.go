package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/yourname/helpdesk/internal/config"
	"github.com/yourname/helpdesk/internal/handler"
	"github.com/yourname/helpdesk/internal/repository"
	"github.com/yourname/helpdesk/internal/router"
	"github.com/yourname/helpdesk/internal/service"
	"github.com/yourname/helpdesk/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	// Repository
	userRepo, err := repository.NewUserRepository(db)
	if err != nil {
		log.Fatalf("user repository: %v", err)
	}

	ticketRepo, err := repository.NewTicketRepository(db)
	if err != nil {
		log.Fatalf("ticket repository: %v", err)
	}

	commentRepo, err := repository.NewCommentRepository(db)
	if err != nil {
		log.Fatalf("comment repository: %v", err)
	}

	// Service
	authService, err := service.NewAuthService(userRepo, cfg.JWTSecret)
	if err != nil {
		log.Fatalf("auth service: %v", err)
	}

	ticketService, err := service.NewTicketService(ticketRepo, userRepo)
	if err != nil {
		log.Fatalf("ticket service: %v", err)
	}

	commentService, err := service.NewCommentService(commentRepo, ticketService)
	if err != nil {
		log.Fatalf("comment service: %v", err)
	}

	userService, err := service.NewUserService(userRepo)
	if err != nil {
		log.Fatalf("user service: %v", err)
	}
	// Handler
	authHandler := handler.NewAuthHandler(authService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	commentHandler := handler.NewCommentHandler(commentService)
	userHandler := handler.NewUserHandler(userService)

	// Router
	engine := router.SetupRouter(authService, authHandler, ticketHandler, commentHandler, userHandler)

	server := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: engine,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	log.Println("server stopped")
}

func openDB(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := migrations.Apply(migrateCtx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	log.Println("database connected")
	return db, nil
}
