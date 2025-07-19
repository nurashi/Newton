package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/nurashi/OpenRouterProject/internal/config"
	"github.com/nurashi/OpenRouterProject/internal/db"
	"github.com/nurashi/OpenRouterProject/internal/telegram"
)

func main() {
	err := godotenv.Load()

	if err != nil { 
		log.Printf("FATAL: .env is not set: %v", err)
	}
	config.Load("config.local")

	if err := db.Connect(); err != nil {
		log.Fatalf("FATAL: db connection failed: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	if err := db.InitSchema(); err != nil {
		log.Fatalf("FATAL: init of schema: %v", err)
	}

	log.Println("Schema inited")

	telegram.RunTelegramBot()
}