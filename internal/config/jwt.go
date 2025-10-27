package config

type Jwt struct {
	SecretToken string `yaml:"secret_token" env-required:"true"`
}
