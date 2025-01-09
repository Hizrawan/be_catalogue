package app

import (
	"be20250107/internal/config"
	"be20250107/internal/modules/authentication"
	"be20250107/internal/modules/cache"
	"be20250107/internal/modules/filestore"
	"be20250107/internal/modules/logger"

	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/nsqio/go-nsq"
)

type Registry struct {
	AppURL string

	config *config.Config
	Config *config.PublicConfig

	Localizer       *Localizer
	DB              *sqlx.DB
	Cache           cache.Cache
	Disks           map[string]filestore.Disk
	Auth            authentication.Auth
	Log             *logger.Logger
	MessageProducer *nsq.Producer
	SigningKey      jwk.RSAPrivateKey
	VerifyKey       jwk.RSAPublicKey
}

func NewRegistry(config *config.Config, appName string) *Registry {
	db, err := NewDatabase(config.Private.Database)
	if err != nil {
		panic(err.Error())
	}

	secretKey, publicKey, err := NewSigningKey(config.Private)
	if err != nil {
		panic(err.Error())
	}

	c, err := NewCache(config.Private.Cache)
	if err != nil {
		panic(err.Error())
	}

	disks, err := NewDisks(config.Private.Storage, secretKey, config.Public)
	if err != nil {
		panic(err.Error())
	}

	authModule := NewAuthModule(config.Private.Auth)
	authModule.Init(db, c, publicKey)

	loggerModule, err := NewLogger(config.Private.Log, config.Public.Debug, appName)
	if err != nil {
		panic(err.Error())
	}

	jwt.Settings(jwt.WithFlattenAudience(true))

	nsqConfig := nsq.NewConfig()
	nsqProducer, err := nsq.NewProducer(config.Public.NsqdHost, nsqConfig)
	if err != nil {
		panic(err.Error())
	}

	localizerModule := NewLocalizer(config.Private.Localizer)

	return &Registry{
		AppURL: config.Public.AppURL,

		config: config,
		Config: config.Public,

		DB:              db,
		Cache:           c,
		Disks:           disks,
		Auth:            authModule,
		Log:             loggerModule,
		Localizer:       localizerModule,
		MessageProducer: nsqProducer,
		SigningKey:      secretKey,
		VerifyKey:       publicKey,
	}
}
