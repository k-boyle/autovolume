package main

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port               string
	DatabaseConnection string `mapstructure:"database_connection"`
}

func LoadConfig(v *viper.Viper) (*Config, error) {
	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.SetEnvPrefix("AV")
	v.AutomaticEnv()

	if err := v.BindEnv("database_connection"); err != nil {
		return nil, err
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
