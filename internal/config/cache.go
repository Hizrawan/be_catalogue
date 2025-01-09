package config

type CacheConfig struct {
	Engine string
	Badger *BadgerConfig
	Redis  *RedisConfig
}

type BadgerConfig struct {
	InMemory   bool `mapstructure:"in_memory"`
	Path       string
	DisableLog bool `mapstructure:"disable_log"`
}

type RedisConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	DBIndex  int `mapstructure:"db_index"`
}
