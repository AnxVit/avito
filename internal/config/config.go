package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env    string `yaml:"env" env-default:"local"`
	DB     `yaml:"db"`
	Server `yaml:"httpServer"`
}

type Server struct {
	Host string `yaml:"host" env-defult:"0.0.0.0"`
	Port string `yaml:"port" env-default:"8082"`

	Timeout time.Duration `yaml:"timeout" env-default:"4s"`
}

type DB struct {
	User     string `yaml:"user" env:"PG_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-required:"true"`
	Host     string `yaml:"host" env:"PG_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"PG_PORT" env-default:"5432"`
	DBName   string `yaml:"dbname" env:"PG_DBNAME" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Println("local.yml not set")
		var cfg Config
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("cannot read env: %s", err)
		}
		return &cfg
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exist", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
