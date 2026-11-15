package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile("config.json", []byte(`{"port": "1234"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AV_DATABASE_CONNECTION", "postgres://test")

	config, err := LoadConfig(viper.New())
	if err != nil {
		t.Fatal(err)
	}

	if config.Port != "1234" {
		t.Errorf("port = %q, want %q", config.Port, "1234")
	}
	if config.DatabaseConnection != "postgres://test" {
		t.Errorf("database connection = %q, want %q", config.DatabaseConnection, "postgres://test")
	}
}
