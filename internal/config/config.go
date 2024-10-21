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
	Env        string `env:"ENV" env-default:"local"`
	Database   DatabaseConfig
	HTTPServer HTTPServerConfig
}

type HTTPServerConfig struct {
	Host string `env:"HTTP_SERVER_HOST" env-default:"localhost"`
	Port int    `env:"HTTP_SERVER_PORT" env-default:"8080"`
}

type DatabaseConfig struct {
	User           string `env:"POSTGRES_USER" env-required:"true"`
	Password       string `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName         string `env:"POSTGRES_DB" env-required:"true"`
	Port           string `env:"POSTGRES_PORT" env-default:"5432"`
	Host           string `env:"POSTGRES_HOST" env-default:"localhost"`
	RepositoryPath string `env:"REPOSITORY_PATH" env-required:"true"`
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

//package config
//
//import (
//	"github.com/ilyakaznacheev/cleanenv"
//)
//
//const (
//	defaultConfigPath = ".env"
//)
//
//// Config хранит конфигурацию приложения
//type Config struct {
//	Env        string `env:"ENV" env-default:"local"`
//	Database   DatabaseConfig
//	HTTPServer HTTPServerConfig
//}
//
//// HTTPServerConfig хранит настройки HTTP сервера
//type HTTPServerConfig struct {
//	Host string `env:"HTTP_SERVER_HOST" env-default:"0.0.0.0"`
//	Port int    `env:"HTTP_SERVER_PORT" env-default:"8080"`
//}
//
//// DatabaseConfig хранит настройки подключения к базе данных
//type DatabaseConfig struct {
//	User           string `env:"POSTGRES_USER" env-default:"user"`
//	Password       string `env:"POSTGRES_PASSWORD" env-default:"password"`
//	DBName         string `env:"POSTGRES_DB" env-default:"aiChallenge"`
//	Port           string `env:"POSTGRES_PORT" env-default:"5432"`
//	Host           string `env:"POSTGRES_HOST" env-default:"db"`
//	RepositoryPath string `env:"REPOSITORY_PATH" env-default:"./db/init.sql"`
//}
//
//// MustLoad загружает конфигурацию
//func MustLoad() *Config {
//	var cfg Config
//
//	// Пытаемся загрузить конфигурацию из файла .env
//	if err := cleanenv.ReadConfig(defaultConfigPath, &cfg); err != nil {
//		// Если файл не найден, просто возвращаем конфигурацию с значениями по умолчанию
//		// Использование cleanenv.Load() для загрузки переменных окружения с их значениями по умолчанию
//		if err := cleanenv.ReadEnv(&cfg); err != nil {
//			panic("failed to read environment variables: " + err.Error())
//		}
//	}
//
//	return &cfg
//}
