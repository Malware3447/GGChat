package config

import "github.com/Malware3447/configo"

type Config struct {
	App        configo.App      `yaml:"app" env-required:"true"`
	DatabasePg configo.Database `yaml:"postgres" env-required:"true"`
}

func (c Config) Env() string {
	return c.App.Env
}
