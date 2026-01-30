package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Global variable for config package
var configPath, pathToEnv string

// init initialize function when application has been started
func init() {
	pflag.StringVar(&configPath, "config", "", "path to config file")
	pflag.StringVar(&pathToEnv, "env", "", "path to env file")
}

// Config structure configuration of the project.
type Config struct {
	LogType            int    `mapstructure:"log_type"`
	ConnectionDBString string `mapstructure:"CONNECTION_STRING"`
	Http               struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"http"`
	TGP struct {
		Secret              string        `mapstructure:"TOKEN_SECRET_KEY_GENERATION"`
		AccessTokenDuration time.Duration `mapstructure:"access_token_duration"`
	} `mapstructure:"tgp"`
}

// MustLoadConfig returns a link to the config struct with parameters for the application if not errors
func MustLoadConfig() *Config {
	if err := fetchConfigPath(); err != nil {
		panic(err)
	}
	config, err := LoadConfigByString(configPath, pathToEnv)
	if err != nil {
		panic(err)
	}
	return config
}

// LoadConfigByString returns link to the config struct with parameters for the application or error if appear
func LoadConfigByString(cPath, ePath string) (*Config, error) {
	var config Config

	if cPath == "" || ePath == "" {
		return nil, fmt.Errorf("config or env path is empty. can not get data")
	}

	yamlViper := viper.New()
	yamlViper.SetConfigFile(cPath)
	yamlViper.SetConfigType("yaml")

	yamlViper.SetDefault("log_type", 1)
	yamlViper.SetDefault("http.host", "127.0.0.1")
	yamlViper.SetDefault("http.port", 8080)
	yamlViper.SetDefault("tgp.access_token_duration", 15*time.Minute)

	// Set viper for parsing custom types
	yamlViper.SetTypeByDefaultValue(true)

	if err := yamlViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %s : %w", cPath, err)
	}

	if err := yamlViper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("could not load config from %s: %w", cPath, err)
	}

	envViper := viper.New()
	envViper.SetConfigFile(ePath)
	envViper.SetConfigType("env")

	if err := envViper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config from %s: %w", ePath, err)
	}

	config.ConnectionDBString = envViper.Get("CONNECTION_STRING").(string)
	config.TGP.Secret = envViper.Get("TOKEN_SECRET_KEY_GENERATION").(string)

	return &config, nil
}

// fetchConfigPath set global string variable of path to the config file (*.yml)
// if error return it
func fetchConfigPath() error {
	pflag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			return fmt.Errorf("config path is required")
		}
	} else if pathToEnv == "" {
		pathToEnv = os.Getenv("ENV_PATH")
		if pathToEnv == "" {
			return fmt.Errorf("env path is required")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config path does not exist")
	} else if _, err = os.Stat(pathToEnv); os.IsNotExist(err) {
		return fmt.Errorf("env path does not exist")
	}

	return nil
}
