package cache

import (
	"errors"
	"time"
)

type Cache interface {
	Has(key string) (bool, error)

	Get(key string) ([]byte, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)
	GetValue(key string, ptr any) (any, error)
	GetKeysWithPrefix(prefix string) ([]string, error)
	GetValuesWithPrefix(prefix string, arrPtr any) (any, error)
	GetItemsWithPrefix(prefix string, mapPtr any) (any, error)

	Put(key string, value []byte, config *Options) error
	PutString(key string, value string, config *Options) error
	PutInt(key string, value int, config *Options) error
	PutValue(key string, value any, config *Options) error

	Delete(key string) error

	Flush() error
	Close() error
}

type Options struct {
	Expiration time.Duration
}

var ErrKeyNotFound = errors.New("cache key not found")
var ErrInvalidValueCast = errors.New("value cannot be casted to the specified type")
