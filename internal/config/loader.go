package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	OpenRouter OpenRouter `mapstructure:"openrouter"`
	Database   Database   `mapstructure:"database"`
	Telegram   Telegram   `mapstructure:"telegram"`
}

type OpenRouter struct {
	Model   string `mapstructure:"model"`
	Referer string `mapstructure:"referer"`
}

type Database struct {
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

func Load(path string) {
	viper.SetConfigName(path)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	if err := viper.Unmarshal(&App); err != nil {
		log.Fatalf("Config unmarshal error: %v", err)
	}

}
