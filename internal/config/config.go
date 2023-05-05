package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   Server
	Postgres Postgres
	APILayer APILayer
}

type Server struct {
	Addr string `env:"SERVER_ADDR"`
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST"`
	Port     string `env:"POSTGRES_PORT"`
	User     string `env:"POSTGRES_USER"`
	Password string `env:"POSTGRES_PASSWORD"`
	DB       string `env:"POSTGRES_DB"`
}

type APILayer struct {
	APIKey string `env:"APILAYER_APIKEY"`
}

func New() (Config, error) {
	var config Config
	err := cleanenv.ReadEnv(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
