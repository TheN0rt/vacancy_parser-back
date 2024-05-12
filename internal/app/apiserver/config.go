package apiserver

import "vacancy-parser/internal/app/store"

type Config struct {
	BindAddr string `toml:"bind_addr" json:"bind_addr"`
	LogLevel string `toml:"log_level" json:"log_level"`
	Store    *store.Config
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":4040",
		LogLevel: "debug",
		Store:    store.NewConfig(),
	}
}
