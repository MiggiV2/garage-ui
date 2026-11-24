package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	ListenAddr string
	AdminURL   string
	AdminToken string
	Username   string
	Password   string
	S3Endpoint string
	S3Region   string
}

func (c Config) BasicAuthEnabled() bool {
	return c.Username != "" && c.Password != ""
}

type garageFile struct {
	S3API struct {
		Region string `toml:"s3_region"`
	} `toml:"s3_api"`
	Admin struct {
		AdminToken string `toml:"admin_token"`
	} `toml:"admin"`
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		ListenAddr: valueOr(getenv("LISTEN_ADDR"), ":8080"),
		AdminURL:   valueOr(getenv("GARAGE_ADMIN_URL"), "http://garage:3903"),
		AdminToken: getenv("GARAGE_ADMIN_TOKEN"),
		Username:   getenv("UI_USERNAME"),
		Password:   getenv("UI_PASSWORD"),
		S3Endpoint: getenv("S3_ENDPOINT"),
		S3Region:   getenv("S3_REGION"),
	}

	var file garageFile
	if path := getenv("GARAGE_CONFIG_FILE"); path != "" {
		if _, err := toml.DecodeFile(path, &file); err != nil {
			return Config{}, fmt.Errorf("read garage config %s: %w", path, err)
		}
	}

	if cfg.AdminToken == "" {
		if getenv("GARAGE_CONFIG_FILE") == "" {
			return Config{}, errors.New("set GARAGE_ADMIN_TOKEN or GARAGE_CONFIG_FILE")
		}
		if file.Admin.AdminToken == "" {
			return Config{}, fmt.Errorf("no [admin] admin_token in %s", getenv("GARAGE_CONFIG_FILE"))
		}
		cfg.AdminToken = file.Admin.AdminToken
	}
	cfg.S3Region = valueOr(cfg.S3Region, valueOr(file.S3API.Region, "garage"))

	if (cfg.Username == "") != (cfg.Password == "") {
		return Config{}, errors.New("UI_USERNAME and UI_PASSWORD must be set together")
	}
	return cfg, nil
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
