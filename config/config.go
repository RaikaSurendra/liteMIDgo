package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	ServiceNow ServiceNowConfig `mapstructure:"servicenow"`
}

type ServerConfig struct {
	Host string     `mapstructure:"host"`
	Port int        `mapstructure:"port"`
	Auth AuthConfig `mapstructure:"auth"`
}

type AuthConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Enabled  bool   `mapstructure:"enabled"`
}

type ServiceNowConfig struct {
	Instance string `mapstructure:"instance"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	UseHTTPS bool   `mapstructure:"use_https"`
	Timeout  int    `mapstructure:"timeout"`
}

func LoadConfig(configPath string) (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configPath != "" {
		viper.AddConfigPath(configPath)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.litemidgo")
	}

	// Set default values
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.auth.enabled", false)
	viper.SetDefault("server.auth.username", "admin")
	viper.SetDefault("server.auth.password", "change-me")
	viper.SetDefault("servicenow.use_https", true)
	viper.SetDefault("servicenow.timeout", 30)

	// Set environment variable bindings
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LITEMIDGO")

	// Bind environment variables to config keys
	viper.BindEnv("servicenow.instance", "SERVICENOW_INSTANCE")
	viper.BindEnv("servicenow.username", "SERVICENOW_USERNAME")
	viper.BindEnv("servicenow.password", "SERVICENOW_PASSWORD")

	// Bind server authentication environment variables
	viper.BindEnv("server.auth.username", "LITEMIDGO_AUTH_USERNAME")
	viper.BindEnv("server.auth.password", "LITEMIDGO_AUTH_PASSWORD")
	viper.BindEnv("server.auth.enabled", "LITEMIDGO_AUTH_ENABLED")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Printf("Config file not found, using defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.ServiceNow.Instance == "" {
		return fmt.Errorf("ServiceNow instance is required. Set SERVICENOW_INSTANCE environment variable or configure in config file")
	}
	if c.ServiceNow.Username == "" {
		return fmt.Errorf("ServiceNow username is required. Set SERVICENOW_USERNAME environment variable or configure in config file")
	}
	if c.ServiceNow.Password == "" {
		return fmt.Errorf("ServiceNow password is required. Set SERVICENOW_PASSWORD environment variable or configure in config file")
	}
	return nil
}
