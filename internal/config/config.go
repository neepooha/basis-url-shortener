package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string  `yaml:"env" env-default:"local"`
	Storage    Storage `yaml:"storage" env-required:"true"`
	HTTPServer `yaml:"http_server"`
	Clients    ClientConfig `yaml:"clients"`
	AppSecret  string       `yaml:"app_secret" env-required:"true" env:"APP_SECRET"`
}

type Storage struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Dbname   string `yaml:"dbname" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

type Client struct {
	Address      string        `yaml:"address" env-default:"localhost:44044"`
	Timeout      time.Duration `yaml:"timeout"`
	RetriesCount int           `yaml:"retriesCount"`
}

type ClientConfig struct {
	SSO Client `yaml:"sso"`
}

// Must - обозначает, что функция либо выполнится, либо вызовет панику
func MustLoad() *Config {
	// loads environment variables from the .env file
	if err := godotenv.Load("config.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	// get configPath from our new env
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if the file exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config file doesn't exist: ", configPath)
	}

	// read config from yaml
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatal("can't read config ", err)
	}

	return &cfg
}
