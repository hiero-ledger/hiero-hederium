package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig() error {
	viper.SetConfigName("config")
	viper.AddConfigPath("./configs")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	return nil
}
