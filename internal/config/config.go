package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

const (
	defaultConfigPath = ".env"
)

type Config struct {
	Env      string `env:"ENV" env-default:"local"`
	Database DatabaseConfig
}

//
//type HTTPServer struct {
//	Address     string        `env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8080"`
//	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
//	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
//	User        string        `env:"HTTP_SERVER_USER" env-required:"true"`
//	Password    string        `env:"HTTP_SERVER_PASSWORD" env-required:"true"`
//}

type DatabaseConfig struct {
	User        string `env:"POSTGRES_USER" env-required:"true"`
	Password    string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName      string `env:"POSTGRES_DB" env-required:"true"`
	Port        string `env:"POSTGRES_PORT" env-default:"5432"`
	Host        string `env:"POSTGRES_HOST" env-default:"localhost"`
	StoragePath string `env:"STORAGE_PATH" env-required:"true"`
}

func MustLoad() *Config {

	configPath := fetchConfigPath()

	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config" + err.Error())
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

	if res == "" {
		res = defaultConfigPath
	}

	return res
}
