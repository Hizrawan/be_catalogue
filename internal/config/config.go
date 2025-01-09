package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

const ServiceName = "HOKISHOP Catalogues"
const DefaultConfigName = ".hoki-catalogues"

var DefaultConfigLocation = []string{"$HOME", "."}

type Config struct {
	Public  *PublicConfig
	Private *PrivateConfig
}

type ListenConfig struct {
	Host      string
	Port      int
	EnableTLS bool `mapstructure:"enable_tls"`
}

type MigrationConfig struct {
	Version         int  `mapstructure:"version"`
	Migrate         bool `mapstructure:"migrate"`
	RollbackOnError bool `mapstructure:"rollback_on_error"`
	AllowDrop       bool `mapstructure:"allow_drop"`
}

type AdminChatConfig struct {
	AutoAssignInterval int `mapstructure:"auto_assign_interval"`
	MaxChatThreshold   int `mapstructure:"max_chat_threshold"`
}

type ThreeSegmentBarcodeConfig struct {
	SmallAmountContractCode string `mapstructure:"small_amount_contract_code"`
	LargeAmountContractCode string `mapstructure:"large_amount_contract_code"`
	DefaultExpiryInHours    int64
	VirtualAccountPrefix    string `mapstructure:"virtual_account_prefix"`
	TimestampMode           string `mapstructure:"timestamp_mode"`
}

type NsqConfig struct {
	NsqdHost       string `mapstructure:"nsqd_host"`
	NSQLookupdHost string `mapstructure:"nsqlookupd_host"`
	MaxConcurrent  int    `mapstructure:"max_concurrent"`
	Consumers      map[string]NsqConsumerConfig
}

type NsqConsumerConfig struct {
	MaxInFlight int `mapstructure:"max_in_flight"`
	MaxAttempt  int `mapstructure:"max_attempt"`
}

type PublicConfig struct {
	Debug                bool
	AppURL               string       `mapstructure:"app_url"`
	PrometheusAPIJobName string       `mapstructure:"prometheus_api_job_name"`
	Listen               ListenConfig `mapstructure:"listen"`

	Migration                         MigrationConfig           `mapstructure:"migration"`
	DBName                            string                    `mapstructure:"-"`
	AttachmentDiskName                string                    `mapstructure:"attachment_disk_name"`
	ThreeSegmentBarcode               ThreeSegmentBarcodeConfig `mapstructure:"three_segment_barcode"`
	AdminChat                         AdminChatConfig           `mapstructure:"admin_chat"`
	UploadScribeMaxAttempt            int                       `mapstructure:"upload_scribe_max_attempt"`
	MaxRadiusNearestStore             int                       `mapstructure:"max_radius_nearest_store"`
	MaxBalanceMutation                float64                   `mapstructure:"max_balance_mutation"`
	MaxOnlineDriverInactiveTimeSecond int                       `mapstructure:"max_online_driver_inactive_time_second"`
	NsqConfig                         `mapstructure:"nsq"`
}

type PrivateConfig struct {
	SigningKey string `mapstructure:"signing_key"`

	Database  DatabaseConfig
	Storage   StorageConfig
	Cache     CacheConfig
	Auth      AuthConfig
	Log       LogConfig
	Localizer LocalizerConfig
	// PushNotif PushNotifConfig `mapstructure:"pushnotif"`
}

func NewConfig(filename string, configPath []string) *Config {
	var config Config

	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	for _, path := range configPath {
		viper.AddConfigPath(path)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	prefix := strings.Replace(strings.ToUpper(ServiceName), " ", "_", -1)
	viper.SetEnvPrefix(prefix)
	fmt.Println(filename)
	fmt.Println(configPath)
	fmt.Println(prefix)
	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("cannot find %s config file in search directory", ServiceName))
		} else {
			panic(fmt.Errorf("config file was found but another error occured: %w", err))
		}
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err.Error())
	}

	ParseStorageConfig(&config)
	ParseAppURL(&config)
	ParseLogConfig(&config)
	ParseMigrationConfig(&config)
	ParseAdminChatConfig(&config)
	config.Public.DBName = config.Private.Database.Name
	return &config
}

func ParseStorageConfig(config *Config) {
	for i, c := range config.Private.Storage.Disks {
		sub := viper.Sub(fmt.Sprintf("private.storage.disks.%d", i))
		switch c.Driver {
		case "gcs":
			config.Private.Storage.Disks[i].GCSDriverConfig = NewGCSDriverConfig(sub)
		case "public", "local":
			config.Private.Storage.Disks[i].LocalDriverConfig = NewLocalDriverConfig(sub)
		}
	}
}

func ParseLogConfig(config *Config) {
	for i, writer := range config.Private.Log.Writers {
		sub := viper.Sub(fmt.Sprintf("private.log.writers.%d", i))
		switch writer.Driver {
		case "rotating-file":
			config.Private.Log.Writers[i].LogRotatingFileWriterConfig = NewRotatingFileWriterConfig(sub)
		case "single-file":
			config.Private.Log.Writers[i].LogSingleFileWriterConfig = NewSingleFileWriterConfig(sub)

		}
	}
}

func ParseMigrationConfig(config *Config) {
	sub := viper.Sub("public.migration")
	config.Public.Migration = NewMigrationConfig(sub)
}

func ParseAdminChatConfig(config *Config) {
	sub := viper.Sub("public.admin_chat")
	config.Public.AdminChat = NewAdminChatConfig(sub)
}

func ParseThreeSegmentBarcodeConfig(config *Config) {
	sub := viper.Sub("public.three_segment_barcode")
	config.Public.ThreeSegmentBarcode = NewThreeSegmentBarcodeConfig(sub)
}

func ParseAppURL(config *Config) {
	var u *url.URL
	pubCfg := config.Public

	if pubCfg.AppURL != "" {
		parsed, err := url.ParseRequestURI(pubCfg.AppURL)
		if err != nil {
			panic(err.Error())
		}
		u = parsed
	} else {
		scheme := "http"
		if pubCfg.Listen.EnableTLS {
			scheme = "https"
		}
		tUrl := fmt.Sprintf("%s://%s:%d", scheme, pubCfg.Listen.Host, pubCfg.Listen.Port)
		parsed, err := url.ParseRequestURI(tUrl)
		if err != nil {
			panic(err.Error())
		}
		u = parsed
	}

	// Listen to localhost (127.0.0.1) by default if no host is defined
	if u.Hostname() == "" {
		u.Host = fmt.Sprintf("%s:%s", "127.0.0.1", u.Port())
	}

	// Remove port information for common protocols (80 for HTTP, 443 for HTTPS)
	if (u.Port() == "80" && u.Scheme == "http") || (u.Port() == "443" && u.Scheme == "https") {
		u.Host = u.Hostname()
	}
	config.Public.AppURL = u.String()
}

func NewGCSDriverConfig(sub *viper.Viper) GCSDriverConfig {
	return GCSDriverConfig{
		KeyFilePath: sub.GetString("key_file_path"),
		KeyFileJSON: sub.GetString("key_file_json"),
		ProjectID:   sub.GetString("project_id"),
		Bucket:      sub.GetString("bucket"),
		PathPrefix:  sub.GetString("path_prefix"),
		Visibility:  sub.GetString("visibility"),
	}
}

func NewLocalDriverConfig(sub *viper.Viper) LocalDriverConfig {
	return LocalDriverConfig{
		Dir:        sub.GetString("dir"),
		PathPrefix: sub.GetString("path_prefix"),
	}
}

func NewMigrationConfig(sub *viper.Viper) MigrationConfig {
	return MigrationConfig{
		Version:         sub.GetInt("version"),
		Migrate:         sub.GetBool("migrate"),
		RollbackOnError: sub.GetBool("rollback_on_error"),
		AllowDrop:       sub.GetBool("allow_drop"),
	}
}

func NewAdminChatConfig(sub *viper.Viper) AdminChatConfig {
	return AdminChatConfig{
		MaxChatThreshold:   sub.GetInt("max_chat_threshold"),
		AutoAssignInterval: sub.GetInt("auto_assign_interval"),
	}
}

func NewThreeSegmentBarcodeConfig(sub *viper.Viper) ThreeSegmentBarcodeConfig {
	return ThreeSegmentBarcodeConfig{
		SmallAmountContractCode: sub.GetString("small_amount_contract_code"),
		LargeAmountContractCode: sub.GetString("large_amount_contract_code"),
		DefaultExpiryInHours:    sub.GetInt64("default_expiry_in_hours"),
		VirtualAccountPrefix:    sub.GetString("virtual_account_prefix"),
		TimestampMode:           sub.GetString("timestamp_mode"),
	}
}

func NewRotatingFileWriterConfig(sub *viper.Viper) LogRotatingFileWriterConfig {
	return LogRotatingFileWriterConfig{
		Filepath: sub.GetString("filepath"),
		Filename: sub.GetString("filename"),
	}
}

func NewSingleFileWriterConfig(sub *viper.Viper) LogSingleFileWriterConfig {
	return LogSingleFileWriterConfig{
		Filepath: sub.GetString("filepath"),
		Filename: sub.GetString("filename"),
	}
}
