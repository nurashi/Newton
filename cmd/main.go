package main

import (
	"github.com/joho/godotenv"
	"github.com/nurashi/internal/telegram"
)

func main() {
	_ = godotenv.Load()
	telegram.RunTelegramBot()
}