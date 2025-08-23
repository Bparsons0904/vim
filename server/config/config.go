package config

import (
	"fmt"
	"server/internal/logger"

	"github.com/spf13/viper"
)

type Config struct {
	GeneralVersion       string `mapstructure:"GENERAL_VERSION"`
	Environment          string `mapstructure:"ENVIRONMENT"`
	ServerPort           int    `mapstructure:"SERVER_PORT"`
	DatabaseDbPath       string `mapstructure:"DB_PATH"`
	DatabaseCacheAddress string `mapstructure:"DB_CACHE_ADDRESS"`
	DatabaseCachePort    int    `mapstructure:"DB_CACHE_PORT"`
	DatabaseCacheReset   int    `mapstructure:"DB_CACHE_RESET"`
	CorsAllowOrigins     string `mapstructure:"CORS_ALLOW_ORIGINS"`
	SecuritySalt         int    `mapstructure:"SECURITY_SALT"`
	SecurityPepper       string `mapstructure:"SECURITY_PEPPER"`
	SecurityJwtSecret    string `mapstructure:"SECURITY_JWT_SECRET"`
	// SessionCookieName    string `mapstructure:"SESSION_COOKIE_NAME"`
}

var ConfigInstance Config

func InitConfig() (Config, error) {
	log := logger.New("config").Function("InitConfig")
	log.Info("Initializing config")

	// Enable automatic environment variable reading first
	viper.AutomaticEnv()

	// Bind environment variables to config keys
	envVars := []string{
		"GENERAL_VERSION", "ENVIRONMENT", "SERVER_PORT", "DB_PATH",
		"DB_CACHE_ADDRESS", "DB_CACHE_PORT", "DB_CACHE_RESET",
		"CORS_ALLOW_ORIGINS", "SECURITY_SALT", "SECURITY_PEPPER", "SECURITY_JWT_SECRET",
	}
	
	for _, env := range envVars {
		viper.BindEnv(env)
	}

	// Check if key environment variables are already set
	envVarsSet := viper.IsSet("SERVER_PORT") && viper.IsSet("SECURITY_JWT_SECRET")

	if envVarsSet {
		log.Info("Environment variables detected, skipping file loading")
	} else {
		log.Info("Environment variables not found, attempting to load from files")
		
		// Load base .env file
		viper.SetConfigFile(".env")
		viper.SetConfigType("env")

		if err := viper.ReadInConfig(); err != nil {
			log.Warn("Could not find .env file", "error", err)
		} else {
			log.Info("Loaded .env file")
		}

		// Load .env.local overrides if it exists
		viper.SetConfigFile(".env.local")
		if err := viper.MergeInConfig(); err != nil {
			log.Debug("No .env.local file found", "error", err)
		} else {
			log.Info("Loaded .env.local overrides")
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return Config{}, log.Err("Fatal error: could not unmarshal config", err)
	}

	log.Info("Successfully initialized config", "config", config)
	err := validateConfig(config, log)
	if err != nil {
		return Config{}, err
	}
	return ConfigInstance, nil
}

func GetConfig() Config {
	return ConfigInstance
}

func validateConfig(config Config, log logger.Logger) error {
	if config.ServerPort <= 0 {
		return log.Err(
			"Fatal error: invalid server port",
			fmt.Errorf("invalid port: %d", config.ServerPort),
			"port", config.ServerPort,
		)
	}

	ConfigInstance = config
	return nil
}
