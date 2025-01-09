package config

type LocalizerConfig struct {
	Directory          string
	SupportedLanguages []string `mapstructure:"supported_languages"`
}
