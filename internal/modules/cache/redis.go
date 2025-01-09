package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"github.com/go-redis/redis/v8"
	"io"
	"reflect"
	"strings"
	"time"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(client *redis.Client, ctx context.Context) Cache {
	return &RedisCache{
		client,
		ctx,
	}
}

func (c RedisCache) Has(key string) (bool, error) {
	r, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return r == 1, nil
}

func (c RedisCache) Get(key string) ([]byte, error) {
	var b []byte
	_, err := c.GetValue(key, &b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c RedisCache) GetString(key string) (string, error) {
	var s string
	_, err := c.GetValue(key, &s)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (c RedisCache) GetInt(key string) (int, error) {
	var i int
	_, err := c.GetValue(key, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c RedisCache) GetValue(key string, ptr any) (any, error) {
	v, err := c.client.Get(c.ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return "", ErrKeyNotFound
	} else if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte(v))
	err = gob.NewDecoder(buf).Decode(ptr)
	return ptr, err
}

func (c RedisCache) GetKeysWithPrefix(prefix string) ([]string, error) {
	var keys []string

	prefix = strings.TrimSpace(prefix)
	if prefix[len(prefix)-1] != '*' {
		prefix = prefix + "*"
	}
	iter := c.client.Scan(c.ctx, 0, prefix, 0).Iterator()
	for iter.Next(c.ctx) {
		keys = append(keys, iter.Val())
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}

	return keys, nil
}

func (c RedisCache) GetValuesWithPrefix(prefix string, arrPtr any) (any, error) {
	keys, err := c.GetKeysWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	pipe := c.client.Pipeline()
	for _, key := range keys {
		pipe.Get(c.ctx, key)
	}
	res, err := pipe.Exec(c.ctx)
	if err != nil {
		return nil, err
	}

	baseType := reflect.TypeOf(arrPtr)
	elemType := baseType
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Slice {
		elemType = elemType.Elem()
	}

	arr := reflect.MakeSlice(reflect.SliceOf(elemType), len(keys), len(keys))
	for i, r := range res {
		cmd := r.(*redis.StringCmd)
		if cmd.Err() != nil {
			continue
		}
		ins := reflect.New(elemType)
		buf := bytes.NewBuffer([]byte(cmd.Val()))
		err = gob.NewDecoder(buf).DecodeValue(ins)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
		arr.Index(i).Set(ins.Elem())
	}

	baseValue := reflect.ValueOf(arrPtr)
	if baseType.Kind() == reflect.Pointer && !baseValue.IsZero() && baseValue.Elem().Kind() == reflect.Slice {
		reflect.ValueOf(arrPtr).Elem().Set(arr)
	}
	return arr.Interface(), nil
}

func (c RedisCache) GetItemsWithPrefix(prefix string, mapPtr any) (any, error) {
	keys, err := c.GetKeysWithPrefix(prefix)
	if err != nil {
		return nil, err
	}

	pipe := c.client.Pipeline()
	for _, key := range keys {
		pipe.Get(c.ctx, key)
	}
	res, err := pipe.Exec(c.ctx)
	if err != nil {
		return nil, err
	}

	baseType := reflect.TypeOf(mapPtr)
	elemType := baseType
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	if elemType.Kind() == reflect.Map {
		elemType = elemType.Elem()
	}
	keyType := reflect.TypeOf("")

	m := reflect.MakeMap(reflect.MapOf(keyType, elemType))
	for i, r := range res {
		cmd := r.(*redis.StringCmd)
		if cmd.Err() != nil {
			continue
		}

		key := keys[i]
		ins := reflect.New(elemType)
		buf := bytes.NewBuffer([]byte(cmd.Val()))
		err = gob.NewDecoder(buf).DecodeValue(ins)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
		m.SetMapIndex(reflect.ValueOf(key), ins.Elem())
	}

	baseValue := reflect.ValueOf(mapPtr)
	if baseType.Kind() == reflect.Pointer && baseValue.Elem().Kind() == reflect.Map {
		reflect.ValueOf(mapPtr).Elem().Set(m)
	}
	return m.Interface(), nil
}

func (c RedisCache) Put(key string, value []byte, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c RedisCache) PutString(key string, value string, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c RedisCache) PutInt(key string, value int, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c RedisCache) PutValue(key string, value any, config *Options) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	if err != nil {
		return err
	}

	exp := time.Duration(0)
	if config != nil {
		exp = config.Expiration
	}
	_, err = c.client.Set(c.ctx, key, buf.String(), exp).Result()
	return err
}

func (c RedisCache) Delete(key string) error {
	res := c.client.Del(c.ctx, key)
	if res.Err() != nil && !errors.Is(res.Err(), redis.Nil) {
		return res.Err()
	}
	return nil
}

func (c RedisCache) Flush() error {
	res := c.client.FlushAll(c.ctx)
	return res.Err()
}

func (c RedisCache) Close() error {
	return c.client.Close()
}
