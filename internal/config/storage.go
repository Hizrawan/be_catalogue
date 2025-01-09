package config

type StorageConfig struct {
	Disks []DiskConfig
}

type DiskConfig struct {
	Name   string
	Driver string
	GCSDriverConfig
	LocalDriverConfig
}

type GCSDriverConfig struct {
	KeyFilePath string `mapstructure:"key_file_path"`
	KeyFileJSON string `mapstructure:"key_file_json"`
	ProjectID   string `mapstructure:"project_id"`
	Bucket      string
	PathPrefix  string `mapstructure:"path_prefix"`
	Visibility  string
}

type LocalDriverConfig struct {
	Dir        string
	PathPrefix string `mapstructure:"path_prefix"`
}
