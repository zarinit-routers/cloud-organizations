package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Load initializes viper with environment and optional config.yml file.
// Also supports common env aliases (PORT, GIN_MODE, JWT_SECURITY_KEY, DB_CONNECTION, RABBITMQ_URL).
func Load() error {
	_ = gotenv.Load() // best-effort .env

	viper.SetEnvPrefix("RAS")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	viper.SetDefault("server.port", 8060)
	viper.SetDefault("server.mode", "release")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("client.address", "*")

	// Rabbit defaults
	viper.SetDefault("rabbitmq.exchange", "domain.events")
	viper.SetDefault("rabbitmq.routing_keys.created", "organization.created")
	viper.SetDefault("rabbitmq.routing_keys.updated", "organization.updated")
	viper.SetDefault("rabbitmq.routing_keys.deleted", "organization.deleted")

	// Try to read config.yml if present
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		log.Debug("config.yml not found or failed to read", "error", err)
	}

	// Aliases for env vars compatibility (unprefixed)
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" && !viper.IsSet("server.port") {
		if n, err := strconv.Atoi(p); err == nil {
			viper.Set("server.port", n)
		}
	}
	if m := strings.TrimSpace(os.Getenv("GIN_MODE")); m != "" && !viper.IsSet("server.mode") {
		viper.Set("server.mode", m)
	}
	if s := strings.TrimSpace(os.Getenv("JWT_SECURITY_KEY")); s != "" && !viper.IsSet("security.jwt.secret") {
		viper.Set("security.jwt.secret", s)
	}
	if cs := strings.TrimSpace(os.Getenv("DB_CONNECTION")); cs != "" && !viper.IsSet("database.connection_string") {
		viper.Set("database.connection_string", cs)
	}
	if url := strings.TrimSpace(os.Getenv("RABBITMQ_URL")); url != "" && !viper.IsSet("rabbitmq.url") {
		viper.Set("rabbitmq.url", url)
	}

	return nil
}
