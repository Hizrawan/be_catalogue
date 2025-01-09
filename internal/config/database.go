package config

type DatabaseConfig struct {
	Type       string
	Name       string
	Username   string
	Password   string
	Host       string
	Port       int
	Query      []string
	DisableLog bool `mapstructure:"disable_log"`
}
