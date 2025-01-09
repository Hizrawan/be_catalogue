package app

import (
	"context"
	"fmt"

	"be20250107/internal/config"
	"be20250107/internal/modules/cache"

	"github.com/dgraph-io/badger/v3"
	"github.com/go-redis/redis/v8"
)

func NewCache(cc config.CacheConfig) (cache.Cache, error) {
	switch cc.Engine {
	case "badger":
		opt := badger.DefaultOptions(cc.Badger.Path)
		if cc.Badger.InMemory {
			opt = opt.WithInMemory(true)
		}
		if cc.Badger.DisableLog {
			opt.Logger = nil
		}
		bDB, err := badger.Open(opt)
		if err != nil {
			panic(err.Error())
		}
		return cache.NewBadgerCache(bDB), nil
	case "redis":
		addr := fmt.Sprintf("%s:%d", cc.Redis.Host, cc.Redis.Port)
		opt := redis.Options{
			Addr:     addr,
			Username: cc.Redis.Username,
			Password: cc.Redis.Password,
			DB:       cc.Redis.DBIndex,
		}
		r := redis.NewClient(&opt)
		ctx := context.Background()
		return cache.NewRedisCache(r, ctx), nil
	}
	return nil, fmt.Errorf("unsupported cache engine: %s", cc.Engine)
}
