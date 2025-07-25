package main

import (
	"github.com/joho/godotenv"
	"github.com/nurashi/Newton/internal/telegram"
)

func main() {
	_ = godotenv.Load()
	telegram.RunTelegramBot()
}