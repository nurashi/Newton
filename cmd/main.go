package main

import (
	"log"
	"os"

	"github.com/nurashi/Newton/internal/config"
	"github.com/nurashi/Newton/internal/database"
	"github.com/nurashi/Newton/internal/repository"
	"github.com/nurashi/Newton/internal/telegram"
	"github.com/nurashi/Newton/migrations"
)

func main() {
	cfg := config.Load("config")
	if cfg == nil {
		log.Fatal("FATAL: failed to load config")
	}

	cfg.Validate()
	log.Printf("CONFIG: %+v", cfg)

	dbpool, err := database.NewPostgresPool(cfg.Database)
	if err != nil {
		log.Fatalf("FATAL: failed to create new Pool: %v", err)
	}
	defer dbpool.Close()

	log.Println("SUCCESSFULLY CONNECTED TO POSTGRES")

	if err := migrations.RunMigrations(dbpool); err != nil {
		log.Printf("WARNING: failed to run mig: %v", err)
	}

	userService := repository.NewUserRepository(dbpool)

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("FATAL: TELEGRAM_BOT_TOKEN not set in .env or environment")
	}

	telegram.RunTelegramBot(userService)
}
