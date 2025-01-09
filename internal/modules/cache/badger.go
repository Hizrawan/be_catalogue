package cache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"reflect"

	"github.com/dgraph-io/badger/v3"
)

type BadgerCache struct {
	db *badger.DB
}

func NewBadgerCache(db *badger.DB) Cache {
	return &BadgerCache{
		db,
	}
}

func (c BadgerCache) Has(key string) (bool, error) {
	available := true
	err := c.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				available = false
			} else {
				return err
			}
		}
		return nil
	})
	if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
		return false, ErrKeyNotFound
	} else if err != nil {
		return false, err
	}
	return available, nil
}

func (c BadgerCache) Get(key string) ([]byte, error) {
	var b []byte
	_, err := c.GetValue(key, &b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c BadgerCache) GetString(key string) (string, error) {
	var s string
	_, err := c.GetValue(key, &s)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (c BadgerCache) GetInt(key string) (int, error) {
	var i int
	_, err := c.GetValue(key, &i)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c BadgerCache) GetValue(key string, ptr any) (any, error) {
	var value []byte
	err := c.db.View(func(txn *badger.Txn) error {
		v, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		value, err = v.ValueCopy(value)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
		return nil, ErrKeyNotFound
	} else if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(value)
	err = gob.NewDecoder(buf).Decode(ptr)
	return ptr, err
}

func (c BadgerCache) GetKeysWithPrefix(prefix string) ([]string, error) {
	var keys []string

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		var k []byte
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			k = it.Item().KeyCopy(k)
			keys = append(keys, string(k))
		}
		return nil
	})

	return keys, err
}

func (c BadgerCache) GetValuesWithPrefix(prefix string, arrPtr any) (any, error) {
	var vBuf [][]byte
	err := c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		var v []byte
		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			v, err := it.Item().ValueCopy(v)
			if err != nil {
				return err
			}
			vBuf = append(vBuf, v)
		}
		return nil
	})
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

	arr := reflect.MakeSlice(reflect.SliceOf(elemType), len(vBuf), len(vBuf))
	for i, out := range vBuf {
		ins := reflect.New(elemType)
		buf := bytes.NewBuffer(out)
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

func (c BadgerCache) GetItemsWithPrefix(prefix string, mapPtr any) (any, error) {
	var keyBuf [][]byte
	var valBuf [][]byte
	err := c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := []byte(prefix)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			var k, v []byte
			k = it.Item().KeyCopy(k)
			v, err := it.Item().ValueCopy(v)
			if err != nil {
				return err
			}
			keyBuf = append(keyBuf, k)
			valBuf = append(valBuf, v)
		}
		return nil
	})
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
	for i, v := range valBuf {
		key := string(keyBuf[i])
		ins := reflect.New(elemType)
		buf := bytes.NewBuffer(v)
		err = gob.NewDecoder(buf).DecodeValue(ins)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
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

func (c BadgerCache) Put(key string, value []byte, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c BadgerCache) PutString(key string, value string, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c BadgerCache) PutInt(key string, value int, config *Options) error {
	return c.PutValue(key, value, config)
}

func (c BadgerCache) PutValue(key string, value any, config *Options) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(value)
	if err != nil {
		return err
	}

	err = c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), buf.Bytes())
		if config != nil && config.Expiration > 0 {
			e = e.WithTTL(config.Expiration)
		}
		err := txn.SetEntry(e)
		return err
	})
	return err
}

func (c BadgerCache) Delete(key string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil
	}
	return err
}

func (c BadgerCache) Flush() error {
	return c.db.DropAll()
}

func (c BadgerCache) Close() error {
	return c.db.Close()
}
