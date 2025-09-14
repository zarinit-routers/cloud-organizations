package config

import (
	"strings"

	"github.com/spf13/viper"
)

func Load() error {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigName("orgs-config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return err
	}
	return nil
}
