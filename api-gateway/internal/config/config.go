package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig
	Supabase     SupabaseConfig
	Database     DatabaseConfig
	Redis        RedisConfig
	Kafka        KafkaConfig
	Orchestrator OrchestratorConfig
	CORS         CORSConfig
}

type ServerConfig struct {
	Port        string `mapstructure:"SERVER_PORT"`
	Environment string `mapstructure:"ENVIRONMENT"`
}

type SupabaseConfig struct {
	JWTSecret string `mapstructure:"SUPABASE_JWT_SECRET"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"DATABASE_URL"`
	MaxConns int32  `mapstructure:"DATABASE_MAX_CONNS"`
}

type RedisConfig struct {
	URL string `mapstructure:"REDIS_URL"`
}

type KafkaConfig struct {
	Brokers []string
	GroupID string `mapstructure:"KAFKA_GROUP_ID"`
}

type OrchestratorConfig struct {
	Enabled bool   `mapstructure:"ORCHESTRATOR_ENABLED"`
	Addr    string `mapstructure:"ORCHESTRATOR_ADDR"`
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// Read .env file if present; ignore missing file
	_ = viper.ReadInConfig()

	// Defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("DATABASE_MAX_CONNS", 20)
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", "localhost:29092")
	viper.SetDefault("KAFKA_GROUP_ID", "api-gateway-ws")
	viper.SetDefault("ORCHESTRATOR_ENABLED", false)
	viper.SetDefault("ORCHESTRATOR_ADDR", "orchestrator:50051")
	viper.SetDefault("ALLOWED_ORIGINS", "http://localhost:3000")

	jwtSecret := viper.GetString("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET is required")
	}

	dbURL := viper.GetString("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	kafkaBrokersRaw := viper.GetString("KAFKA_BROKERS")
	kafkaBrokers := splitCSV(kafkaBrokersRaw)

	allowedOriginsRaw := viper.GetString("ALLOWED_ORIGINS")
	allowedOrigins := splitCSV(allowedOriginsRaw)

	cfg := &Config{
		Server: ServerConfig{
			Port:        viper.GetString("SERVER_PORT"),
			Environment: viper.GetString("ENVIRONMENT"),
		},
		Supabase: SupabaseConfig{
			JWTSecret: jwtSecret,
		},
		Database: DatabaseConfig{
			URL:      dbURL,
			MaxConns: viper.GetInt32("DATABASE_MAX_CONNS"),
		},
		Redis: RedisConfig{
			URL: viper.GetString("REDIS_URL"),
		},
		Kafka: KafkaConfig{
			Brokers: kafkaBrokers,
			GroupID: viper.GetString("KAFKA_GROUP_ID"),
		},
		Orchestrator: OrchestratorConfig{
			Enabled: viper.GetBool("ORCHESTRATOR_ENABLED"),
			Addr:    viper.GetString("ORCHESTRATOR_ADDR"),
		},
		CORS: CORSConfig{
			AllowedOrigins: allowedOrigins,
		},
	}

	return cfg, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
