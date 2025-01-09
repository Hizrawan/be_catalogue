package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"reflect"
	"strings"
	"time"
)

type InMemoryCache struct {
	state map[string]*cacheEntry
}

type cacheEntry struct {
	Data       []byte
	Expiration time.Time
}

func NewInMemoryCache() Cache {
	return &InMemoryCache{
		state: map[string]*cacheEntry{},
	}
}

func (c *InMemoryCache) Has(key string) (bool, error) {
	v, ok := c.state[key]
	if !ok {
		return false, nil
	}
	if !v.Expiration.IsZero() && v.Expiration.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

func (c *InMemoryCache) Get(key string) ([]byte, error) {
	var b []byte
	_, err := c.GetValue(key, &b)
	if err != nil {
		return nil, err
	}
	return b, err
}

func (c *InMemoryCache) GetString(key string) (string, error) {
	var s string
	_, err := c.GetValue(key, &s)
	if err != nil {
		return "", err
	}
	return s, err
}

func (c *InMemoryCache) GetInt(key string) (int, error) {
	var i int
	_, err := c.GetValue(key, &i)
	if err != nil {
		return 0, err
	}
	return i, err
}

func (c *InMemoryCache) GetValue(key string, ptr any) (any, error) {
	v, ok := c.state[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	if !v.Expiration.IsZero() && v.Expiration.Before(time.Now()) {
		return nil, ErrKeyNotFound
	}

	buf := bytes.NewBuffer(v.Data)
	err := gob.NewDecoder(buf).Decode(ptr)
	return ptr, err
}

func (c *InMemoryCache) GetKeysWithPrefix(prefix string) ([]string, error) {
	l := make([]string, 0)
	for k, v := range c.state {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if !v.Expiration.IsZero() && v.Expiration.Before(time.Now()) {
			continue
		}
		l = append(l, k)
	}
	return l, nil
}

func (c *InMemoryCache) GetValuesWithPrefix(prefix string, arrPtr any) (any, error) {
	keys, err := c.GetKeysWithPrefix(prefix)
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
	for i, k := range keys {
		v := c.state[k]

		ins := reflect.New(elemType)
		buf := bytes.NewBuffer(v.Data)
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

func (c *InMemoryCache) GetItemsWithPrefix(prefix string, mapPtr any) (any, error) {
	keys, err := c.GetKeysWithPrefix(prefix)
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
	for _, k := range keys {
		v := c.state[k]

		ins := reflect.New(elemType)
		buf := bytes.NewBuffer(v.Data)
		err = gob.NewDecoder(buf).DecodeValue(ins)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
		m.SetMapIndex(reflect.ValueOf(k), ins.Elem())
	}

	baseValue := reflect.ValueOf(mapPtr)
	if baseType.Kind() == reflect.Pointer && baseValue.Elem().Kind() == reflect.Map {
		reflect.ValueOf(mapPtr).Elem().Set(m)
	}
	return m.Interface(), nil
}

func (c *InMemoryCache) Put(key string, value []byte, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c *InMemoryCache) PutString(key string, value string, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c *InMemoryCache) PutInt(key string, value int, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c *InMemoryCache) PutValue(key string, value any, config *Options) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	if err != nil {
		return err
	}

	exp := time.Time{}
	if config != nil {
		exp = time.Now().Add(config.Expiration)
	}
	c.state[key] = &cacheEntry{
		Data:       buf.Bytes(),
		Expiration: exp,
	}
	return nil
}

func (c *InMemoryCache) Delete(key string) error {
	if exist, err := c.Has(key); err != nil {
		return err
	} else if !exist {
		return ErrKeyNotFound
	}
	delete(c.state, key)
	return nil
}

func (c *InMemoryCache) PruneExpiredKeys() error {
	var prune []string
	for k, v := range c.state {
		if !v.Expiration.IsZero() && v.Expiration.Before(time.Now()) {
			prune = append(prune, k)
		}
	}
	for _, k := range prune {
		delete(c.state, k)
	}
	return nil
}

func (c *InMemoryCache) Flush() error {
	for k := range c.state {
		delete(c.state, k)
	}
	return nil
}

func (c *InMemoryCache) Close() error {
	return nil
}
