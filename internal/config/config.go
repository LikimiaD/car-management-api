package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"path/filepath"
	"time"
)

type DatabaseConfig struct {
	Name     string `env:"DB_NAME"     env-required:"true"`
	User     string `env:"DB_USER"     env-required:"true"`
	Password string `env:"DB_PASSWORD" env-required:"true"`
	Port     string `env:"DB_PORT"     env-required:"true"`
	Host     string `env:"DB_HOST"     env-required:"true"`
}

type HTTPServer struct {
	Address          string        `env:"HTTP_ADDRESS"             env-default:"0.0.0.0:8080"`
	Timeout          time.Duration `env:"HTTP_TIMEOUT"             env-default:"5s"`
	IdleTimeout      time.Duration `env:"HTTP_IDLE_TIMEOUT"        env-default:"30s"`
	MaxWorkers       int           `env:"HTTP_MAX_WORKERS"         env-default:"10"`
	ThirdPartyAPIURL string        `env:"HTTP_THIRD_PARTY_API_URL" env-required:"true"`
	DebugMode        bool          `env:"HTTP_DEBUG_MODE"          env-default:"true"`
}

type Config struct {
	SecretKey      string `env:"SECRET_KEY"`
	HTTPServer     `env:"http_server"`
	DatabaseConfig `env:"database"`
}

func GetConfig() *Config {
	defer func(start time.Time) {
		fmt.Printf("%s [%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), "START", "load config", time.Since(start))
	}(time.Now())
	return loadConfig()
}

func loadConfig() *Config {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("error getting executable path: %s", err)
	}

	exeDir := filepath.Dir(exePath)

	configPath := filepath.Join(exeDir, "docs", ".env")
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	err = cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
