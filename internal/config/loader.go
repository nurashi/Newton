package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	OpenRouter OpenRouter `mapstructure:"openrouter"`
	Database   PostgreSQL `mapstructure:"database"`
	Telegram   Telegram   `mapstructure:"telegram"`
}

type OpenRouter struct {
	Model   string `mapstructure:"model"`
	Referer string `mapstructure:"referer"`
}

type PostgreSQL struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type Telegram struct {
	Token string `mapstructure:"token"`
}

var App Config

func Load(path string) *Config {
	_ = godotenv.Load()

	viper.SetConfigName(path)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.BindEnv("telegram.token", "TELEGRAM_BOT_TOKEN")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("openrouter.model", "OPENROUTER_API_KEY") 
	
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	if err := viper.Unmarshal(&App); err != nil {
		log.Fatalf("Config unmarshal error: %v", err)
	}
	return &App
}


func (c *Config) Validate() {
	if c.Telegram.Token == "" {
		log.Fatal("Missing Telegram token in config")
	}
	if c.Database.User == "" || c.Database.Password == "" || c.Database.Name == "" {
		log.Fatal("Database credentials are incomplete in config")
	}
	if c.OpenRouter.Model == "" {
		log.Fatal("Missing OpenRouter model in config")
	}
}
