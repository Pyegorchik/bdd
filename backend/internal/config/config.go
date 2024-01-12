package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Postgres     *PostgresConfig
		Server       *ServerConfig
		Handler      *HandlerConfig
		Service      *ServiceConfig
		TokenManager *TokenManagerConfig
	}

	PostgresConfig struct {
		Host     string
		User     string
		Password string
		DBName   string
		Port     int
	}

	ServerConfig struct {
		Port           int
		ReadTimeout    time.Duration
		WriteTimeout   time.Duration
		MaxHeaderBytes int
	}

	HandlerConfig struct {
		RequestTimeout time.Duration
		SwaggerHost    string
	}

	ServiceConfig struct {
		AccessTokenTTL  time.Duration
		RefreshTokenTTL time.Duration
		StaticPath string
	}

	TokenManagerConfig struct {
		SigningKey string
	}
)

func Init(configPath string) (*Config, error) {
	jsonCfg := viper.New()
	jsonCfg.AddConfigPath(filepath.Dir(configPath))
	jsonCfg.SetConfigName(filepath.Base(configPath))

	if err := jsonCfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config/Init/jsonCfg.ReadInConfig: %w", err)
	}

	envCfg := viper.New()
	envCfg.SetConfigFile(".env")

	if err := envCfg.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config/Init/envCfg.ReadInConfig: %w", err)
	}

	return &Config{
		Postgres: &PostgresConfig{
			Host:     envCfg.GetString("POSTGRES_HOST"),
			User:     envCfg.GetString("POSTGRES_USER"),
			Password: envCfg.GetString("POSTGRES_PASSWORD"),
			DBName:   envCfg.GetString("POSTGRES_DB"),
			Port:     envCfg.GetInt("POSTGRES_PORT"),
		},
		Server: &ServerConfig{
			Port:           jsonCfg.GetInt("server.port"),
			ReadTimeout:    jsonCfg.GetDuration("server.readTimeout"),
			WriteTimeout:   jsonCfg.GetDuration("server.writeTimeout"),
			MaxHeaderBytes: jsonCfg.GetInt("server.maxHeaderBytes"),
		},
		Handler: &HandlerConfig{
			RequestTimeout: jsonCfg.GetDuration("handler.requestTimeout"),
			SwaggerHost:    jsonCfg.GetString("handler.swaggerHost"),
		},
		Service: &ServiceConfig{
			AccessTokenTTL:  jsonCfg.GetDuration("service.accessTTL"),
			RefreshTokenTTL: jsonCfg.GetDuration("service.refreshTTL"),
			StaticPath: jsonCfg.GetString("service.staticPath"),
		},
		TokenManager: &TokenManagerConfig{
			SigningKey: envCfg.GetString("JWT_SIGNING_KEY"),
		},
	}, nil
}

func (p *PostgresConfig) PgSource() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.DBName)
}