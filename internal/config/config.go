package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env         string `yaml:"env" env-required:"true"`
	Migrations  `yaml:"migrations"`
	Auth        `yaml:"auth"`
	HttpServer  `yaml:"http_server"`
	EmailSender `yaml:"email_sender"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	SecretKey   string        `yaml:"secret_key" env-required:"true"`
}

type Auth struct {
	LinkTtl  time.Duration `yaml:"link_ttl" env-required:"true"`
	TokenTtl time.Duration `yaml:"token_ttl" env-required:"true"`
}

type EmailSender struct {
	ApiKey string `yaml:"api_key" env-required:"true"`
	Name   string `yaml:"name" env-required:"true"`
	Email  string `yaml:"email" env-required:"true"`
}

type Migrations struct {
	Path string `yaml:"path"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()

	if configPath == "" {
		panic("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file does not exist: %s", configPath))
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("cannot read config: %s", err))
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
