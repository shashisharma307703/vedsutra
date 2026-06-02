package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database PoolConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port            string        `env:"SERVER_PORT" default:"8080"`
	Mode            string        `env:"GIN_MODE" default:"debug"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" default:"30s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" default:"30s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" default:"5s"`
}

type LogConfig struct {
	Level  string `env:"LOG_LEVEL" default:"info"`
	Format string `env:"LOG_FORMAT" default:"json"`
}

type PoolConfig struct {
	Host     string `env:"PGPOOL_HOST" default:"localhost"`
	Port     string `env:"PGPOOL_PORT" default:"5432"`
	User     string `env:"DB_USER" default:"sms_admin"`
	Password string `env:"DB_PASSWORD" default:"123456"`
	Database string `env:"DB_NAME" default:"school_erp"`

	MaxConnections    int32         `env:"APP_MAX_CONN" default:"5"`
	MinConnections    int32         `env:"APP_MIN_CONN" default:"2"`
	ConnMaxLifetime   time.Duration `env:"APP_CONN_MAX_LIFETIME" default:"15m"`
	ConnMaxIdleTime   time.Duration `env:"APP_CONN_MAX_IDLE" default:"5m"`
	AcquireTimeout    time.Duration `env:"APP_ACQUIRE_TIMEOUT" default:"5s"`
	ConnectionTimeout time.Duration `env:"APP_CONNECTION_TIMEOUT" default:"10s"`
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", ":8080"),
			Mode:            getEnv("GIN_MODE", "debug"),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second),
		},
		Database: PoolConfig{
			Host:              getEnv("PGPOOL_HOST", "localhost"),
			Port:              getEnv("PGPOOL_PORT", "5432"),
			User:              getEnv("DB_USER", "sms_admin"),
			Password:          getEnv("DB_PASSWORD", "123456"),
			Database:          getEnv("DB_NAME", "school_erp"),
			MaxConnections:    int32(getEnvInt("APP_MAX_CONN", 5)),
			MinConnections:    int32(getEnvInt("APP_MIN_CONN", 2)),
			ConnMaxLifetime:   getEnvDuration("APP_CONN_MAX_LIFETIME", 15*time.Minute),
			ConnMaxIdleTime:   getEnvDuration("APP_CONN_MAX_IDLE", 5*time.Minute),
			AcquireTimeout:    getEnvDuration("APP_ACQUIRE_TIMEOUT", 5*time.Second),
			ConnectionTimeout: getEnvDuration("APP_CONNECTION_TIMEOUT", 10*time.Second),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

func GetEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}

func getEnv(k, d string) string     { return GetEnv(k, d) }
func getEnvInt(k string, d int) int { return GetEnvInt(k, d) }
func getEnvDuration(k string, d time.Duration) time.Duration {
	return GetEnvDuration(k, d)
}
