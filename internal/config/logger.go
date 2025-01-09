package config

type LogConfig struct {
	Writers []LogWriterConfig
}

type LogWriterConfig struct {
	Name     string
	Driver   string
	LogLevel []string `mapstructure:"log_level"`
	LogRotatingFileWriterConfig
	LogSingleFileWriterConfig
}

type LogRotatingFileWriterConfig struct {
	Filepath string
	Filename string
}

type LogSingleFileWriterConfig struct {
	Filepath string
	Filename string
}
