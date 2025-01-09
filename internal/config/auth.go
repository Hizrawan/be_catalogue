package config

type XinchuanAuthConfig struct {
	ID       int
	Secret   string
	BaseURL  string `mapstructure:"base_url"`
	Callback string
}

type MobileBEAuthConfig struct {
	Secret  string
	BaseURL string `mapstructure:"base_url"`
}

type AuthConfig struct {
	XinchuanAuth XinchuanAuthConfig `mapstructure:"xinchuan_auth"`
	MobileBEAuth MobileBEAuthConfig `mapstructure:"mobile_be_auth"`
}
