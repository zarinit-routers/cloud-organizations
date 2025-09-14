package logger

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// Setup configures the global logger according to config and service name.
// It keeps things minimal: set the level from viper key log.level (debug/info/warning/error).
// The service name can be used as a prefix field when logging manually; here we just set level.
func Setup(service string) {
	lvl := strings.ToLower(strings.TrimSpace(viper.GetString("log.level")))
	switch lvl {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info", "":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	case "error", "err":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
