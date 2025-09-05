package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for our application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	Logger    LoggerConfig
	Tenant    TenantConfig
	Security  SecurityConfig
	Features  FeatureConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationsPath  string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey           string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	Issuer              string
	Audience            string
	IncludeTenantClaims bool
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level          string
	Format         string
	IncludeTenantID bool
}

// TenantConfig holds tenant-specific configuration
type TenantConfig struct {
	DefaultTrialDays    int
	MaxTenantsPerDomain int
	SlugMinLength       int
	SlugMaxLength       int
	AllowSubdomains     bool
	RequireSSL          bool
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	PasswordMinLength     int
	PasswordRequireUpper  bool
	PasswordRequireLower  bool
	PasswordRequireDigit  bool
	PasswordRequireSymbol bool
	MaxLoginAttempts      int
	LockoutDuration       time.Duration
	SessionTimeout        time.Duration
}

// FeatureConfig holds feature flag configuration
type FeatureConfig struct {
	EnableMultiTenancy     bool
	EnableSubscriptions    bool
	EnableUsageLimits      bool
	EnableTrialPeriods     bool
	EnableSubdomains       bool
	EnableCustomDomains    bool
	EnableFeatureGating    bool
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			DBName:          getEnv("DB_NAME", "adol_pos"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "migrations"),
		},
		JWT: JWTConfig{
			SecretKey:           getEnv("JWT_SECRET_KEY", "your-256-bit-secret"),
			AccessTokenExpiry:   getDurationEnv("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry:  getDurationEnv("JWT_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
			Issuer:              getEnv("JWT_ISSUER", "adol-pos"),
			Audience:            getEnv("JWT_AUDIENCE", "adol-pos-api"),
			IncludeTenantClaims: getBoolEnv("JWT_INCLUDE_TENANT_CLAIMS", true),
		},
		Logger: LoggerConfig{
			Level:           getEnv("LOG_LEVEL", "info"),
			Format:          getEnv("LOG_FORMAT", "json"),
			IncludeTenantID: getBoolEnv("LOG_INCLUDE_TENANT_ID", true),
		},
		Tenant: TenantConfig{
			DefaultTrialDays:    getIntEnv("TENANT_DEFAULT_TRIAL_DAYS", 30),
			MaxTenantsPerDomain: getIntEnv("TENANT_MAX_PER_DOMAIN", 1),
			SlugMinLength:       getIntEnv("TENANT_SLUG_MIN_LENGTH", 3),
			SlugMaxLength:       getIntEnv("TENANT_SLUG_MAX_LENGTH", 63),
			AllowSubdomains:     getBoolEnv("TENANT_ALLOW_SUBDOMAINS", true),
			RequireSSL:          getBoolEnv("TENANT_REQUIRE_SSL", false),
		},
		Security: SecurityConfig{
			PasswordMinLength:     getIntEnv("SECURITY_PASSWORD_MIN_LENGTH", 8),
			PasswordRequireUpper:  getBoolEnv("SECURITY_PASSWORD_REQUIRE_UPPER", true),
			PasswordRequireLower:  getBoolEnv("SECURITY_PASSWORD_REQUIRE_LOWER", true),
			PasswordRequireDigit:  getBoolEnv("SECURITY_PASSWORD_REQUIRE_DIGIT", true),
			PasswordRequireSymbol: getBoolEnv("SECURITY_PASSWORD_REQUIRE_SYMBOL", false),
			MaxLoginAttempts:      getIntEnv("SECURITY_MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:       getDurationEnv("SECURITY_LOCKOUT_DURATION", 15*time.Minute),
			SessionTimeout:        getDurationEnv("SECURITY_SESSION_TIMEOUT", 8*time.Hour),
		},
		Features: FeatureConfig{
			EnableMultiTenancy:  getBoolEnv("FEATURE_ENABLE_MULTI_TENANCY", true),
			EnableSubscriptions: getBoolEnv("FEATURE_ENABLE_SUBSCRIPTIONS", true),
			EnableUsageLimits:   getBoolEnv("FEATURE_ENABLE_USAGE_LIMITS", true),
			EnableTrialPeriods:  getBoolEnv("FEATURE_ENABLE_TRIAL_PERIODS", true),
			EnableSubdomains:    getBoolEnv("FEATURE_ENABLE_SUBDOMAINS", true),
			EnableCustomDomains: getBoolEnv("FEATURE_ENABLE_CUSTOM_DOMAINS", false),
			EnableFeatureGating: getBoolEnv("FEATURE_ENABLE_FEATURE_GATING", true),
		},
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv gets an environment variable as integer or returns a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getDurationEnv gets an environment variable as duration or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getBoolEnv gets an environment variable as boolean or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetDatabaseURL returns the database connection URL
func (c *Config) GetDatabaseURL() string {
	return "postgres://" + c.Database.User + ":" + c.Database.Password + "@" + c.Database.Host + ":" + c.Database.Port + "/" + c.Database.DBName + "?sslmode=" + c.Database.SSLMode
}

// IsMultiTenancyEnabled returns whether multi-tenancy is enabled
func (c *Config) IsMultiTenancyEnabled() bool {
	return c.Features.EnableMultiTenancy
}

// IsSubscriptionsEnabled returns whether subscriptions are enabled
func (c *Config) IsSubscriptionsEnabled() bool {
	return c.Features.EnableSubscriptions
}

// IsFeatureGatingEnabled returns whether feature gating is enabled
func (c *Config) IsFeatureGatingEnabled() bool {
	return c.Features.EnableFeatureGating
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.SecretKey == "" || c.JWT.SecretKey == "your-256-bit-secret" {
		return fmt.Errorf("JWT secret key must be set and not use default value")
	}
	
	if len(c.JWT.SecretKey) < 32 {
		return fmt.Errorf("JWT secret key must be at least 32 characters long")
	}
	
	if c.Database.Host == "" {
		return fmt.Errorf("database host must be set")
	}
	
	if c.Database.DBName == "" {
		return fmt.Errorf("database name must be set")
	}
	
	if c.Tenant.SlugMinLength < 1 {
		return fmt.Errorf("tenant slug minimum length must be at least 1")
	}
	
	if c.Tenant.SlugMaxLength > 63 {
		return fmt.Errorf("tenant slug maximum length cannot exceed 63 characters")
	}
	
	if c.Tenant.SlugMinLength >= c.Tenant.SlugMaxLength {
		return fmt.Errorf("tenant slug minimum length must be less than maximum length")
	}
	
	if c.Security.PasswordMinLength < 8 {
		return fmt.Errorf("password minimum length must be at least 8")
	}
	
	if c.Security.MaxLoginAttempts < 1 {
		return fmt.Errorf("max login attempts must be at least 1")
	}
	
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, strings.ToLower(c.Logger.Level)) {
		return fmt.Errorf("invalid log level: %s, must be one of: %s", c.Logger.Level, strings.Join(validLogLevels, ", "))
	}
	
	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, strings.ToLower(c.Logger.Format)) {
		return fmt.Errorf("invalid log format: %s, must be one of: %s", c.Logger.Format, strings.Join(validLogFormats, ", "))
	}
	
	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}