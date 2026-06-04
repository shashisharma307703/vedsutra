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
	Auth     AuthConfig
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

type AuthConfig struct {
	// JWT configuration
	JWTSecret              string        `env:"JWT_SECRET"`
	JWTAccessTokenExpiry   time.Duration `env:"JWT_ACCESS_TOKEN_EXPIRY" default:"15m"`
	JWTRefreshTokenExpiry  time.Duration `env:"JWT_REFRESH_TOKEN_EXPIRY" default:"168h"`
	JWTAlgorithm           string        `env:"JWT_ALGORITHM" default:"HS256"`

	// OIDC configuration
	OIDCDiscoveryURL       string        `env:"OIDC_DISCOVERY_URL"`
	OIDCClientID           string        `env:"OIDC_CLIENT_ID"`
	OIDCClientSecret       string        `env:"OIDC_CLIENT_SECRET"`
	OIDCScopes             string        `env:"OIDC_SCOPES" default:"openid profile email roles"`
	OIDCDiscoveryCacheTTL  time.Duration `env:"OIDC_DISCOVERY_CACHE_TTL" default:"1h"`

	// Local auth configuration
	LocalAuthEnabled       bool   `env:"LOCAL_AUTH_ENABLED" default:"true"`
	PasswordMinLength      int    `env:"PASSWORD_MIN_LENGTH" default:"8"`
	BcryptCost             int    `env:"BCRYPT_COST" default:"10"`

	// Refresh token configuration
	RefreshTokenRotation   bool          `env:"REFRESH_TOKEN_ROTATION" default:"true"`
	RefreshTokenMaxAge     time.Duration `env:"REFRESH_TOKEN_MAX_AGE" default:"720h"`

	// Security configuration
	AllowedOrigins         string `env:"ALLOWED_ORIGINS" default:"localhost:3000"`
	CookieSecure           bool   `env:"COOKIE_SECURE" default:"false"`
	CookieSameSite         string `env:"COOKIE_SAME_SITE" default:"Strict"`

	// Session configuration
	SessionTimeout         time.Duration `env:"SESSION_TIMEOUT" default:"24h"`
	SessionCookieName      string        `env:"SESSION_COOKIE_NAME" default:"vedsutra_session"`
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
		Auth: AuthConfig{
			JWTSecret:             getEnv("JWT_SECRET", ""),
			JWTAccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			JWTRefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			JWTAlgorithm:          getEnv("JWT_ALGORITHM", "HS256"),
			OIDCDiscoveryURL:      getEnv("OIDC_DISCOVERY_URL", ""),
			OIDCClientID:          getEnv("OIDC_CLIENT_ID", ""),
			OIDCClientSecret:      getEnv("OIDC_CLIENT_SECRET", ""),
			OIDCScopes:            getEnv("OIDC_SCOPES", "openid profile email roles"),
			OIDCDiscoveryCacheTTL: getEnvDuration("OIDC_DISCOVERY_CACHE_TTL", 1*time.Hour),
			LocalAuthEnabled:      getEnvBool("LOCAL_AUTH_ENABLED", true),
			PasswordMinLength:     getEnvInt("PASSWORD_MIN_LENGTH", 8),
			BcryptCost:            getEnvInt("BCRYPT_COST", 10),
			RefreshTokenRotation:  getEnvBool("REFRESH_TOKEN_ROTATION", true),
			RefreshTokenMaxAge:    getEnvDuration("REFRESH_TOKEN_MAX_AGE", 30*24*time.Hour),
			AllowedOrigins:        getEnv("ALLOWED_ORIGINS", "localhost:3000"),
			CookieSecure:          getEnvBool("COOKIE_SECURE", false),
			CookieSameSite:        getEnv("COOKIE_SAME_SITE", "Strict"),
			SessionTimeout:        getEnvDuration("SESSION_TIMEOUT", 24*time.Hour),
			SessionCookieName:     getEnv("SESSION_COOKIE_NAME", "vedsutra_session"),
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

func GetEnvBool(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		return v == "true" || v == "1" || v == "yes"
	}
	return defaultValue
}

func getEnv(k, d string) string     { return GetEnv(k, d) }
func getEnvInt(k string, d int) int { return GetEnvInt(k, d) }
func getEnvDuration(k string, d time.Duration) time.Duration {
	return GetEnvDuration(k, d)
}
func getEnvBool(k string, d bool) bool {
	return GetEnvBool(k, d)
}
