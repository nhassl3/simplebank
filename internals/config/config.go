package config

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}

type Config struct {
	EnvDefault         int    `yaml:"env" env-required:"true"`
	ConnectionDBString string `yaml:"connection_db" env-required:"true"`
	Http               struct {
		Host string `yaml:"host" env-default:"localhost"`
		Port int    `yaml:"port" env-default:"8080"`
	} `yaml:"http"`
}

func MustLoadConfig() *Config {
	if err := fetchConfigPath(); err != nil {
		panic(err)
	}
	return MustLoadConfigByString(configPath)
}

func MustLoadConfigByString(path string) *Config {
	var config Config

	if err := cleanenv.ReadConfig(path, &config); err != nil {
		panic("Could not load config from " + path)
	}
	config.ConnectionDBString = strings.Replace(
		config.ConnectionDBString,
		"${DB_PASSWORD}",
		os.Getenv("DATABASE_PASSWORD"),
		-1,
	)

	return &config
}

func fetchConfigPath() error {
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			return errors.New("config path is required")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.New("config path does not exist")
	}

	return nil
}
