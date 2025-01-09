package ratelimiter

import (
	"errors"
	"fmt"
	"time"

	"be20250107/internal/modules/cache"
)

var ErrRateLimited = errors.New("request failed due to rate-limit")

type RateLimiter struct {
	cache      cache.Cache
	Expiration time.Duration
	Namespace  string
	IDs        []string
	KeyPrefix  string
}

func New(cache cache.Cache, namespace string, ids []string, expiration time.Duration) RateLimiter {
	joinedID := ""
	for _, i := range ids {
		if joinedID == "" {
			joinedID = i
		} else {
			joinedID = joinedID + "_" + i
		}
	}
	prefix := fmt.Sprintf("rl:%s:%s", namespace, joinedID)
	return RateLimiter{
		cache:      cache,
		Namespace:  namespace,
		IDs:        ids,
		Expiration: expiration,
		KeyPrefix:  prefix,
	}
}

func (l *RateLimiter) Count() (int, error) {
	attempts, err := l.cache.GetKeysWithPrefix(l.KeyPrefix)
	if err != nil {
		return -1, nil
	}
	return len(attempts), nil
}

func (l *RateLimiter) Guard(max int) error {
	count, err := l.Count()
	if err != nil {
		return err
	}
	if count >= max {
		return ErrRateLimited
	}
	return nil
}

func (l *RateLimiter) RecordAttempt() error {
	now := time.Now().Unix()
	key := fmt.Sprintf("%s:%d", l.KeyPrefix, now)
	opt := &cache.Options{Expiration: l.Expiration}
	if l.Expiration <= 0 {
		opt = nil
	}
	return l.cache.PutValue(key, now, opt)
}

func (l *RateLimiter) Clear() error {
	keys, err := l.cache.GetKeysWithPrefix(l.KeyPrefix)
	if err != nil {
		return err
	}
	for _, k := range keys {
		if err = l.cache.Delete(k); err != nil {
			return err
		}
	}
	return nil
}
