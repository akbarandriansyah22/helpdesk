package main

import (
	"context"
	"log"
	"time"

	"github.com/yourname/helpdesk/internal/config"
	"github.com/yourname/helpdesk/internal/database"
	"github.com/yourname/helpdesk/migrations"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.Apply(ctx, db); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	log.Println("migrations and seed applied")
}
