package utils

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"sync"
)

type RedisCache struct {
	client    redis.Conn
	keyPrefix RedisKey
	maxAge    int
	mux       sync.Mutex
}

type sourceFunc func(string) (interface{}, error)

func NewRedisCache(client redis.Conn, keyPrefix RedisKey, maxAge int) *RedisCache {
	redisCache := &RedisCache{
		client:    client,
		keyPrefix: keyPrefix,
		maxAge:    maxAge,
	}

	return redisCache
}

func (cache *RedisCache) Query(key RedisKey, index string, source sourceFunc) (interface{}, error) {
	cache.Lock()
	defer cache.Unlock()

	reply, err := cache.client.Do("GET", cache.keyPrefix.AppendKey(key).Append(index))

	var data interface{}
	if err != nil {
		log.Debug("RedisCache: could not query cache" + err.Error())
		return nil, errors.New("RedisCache: could not query cache" + err.Error())
	}

	if reply != nil {
		log.Debug("RedisCache: Found index in cache")
		err = cache.deserialize(reply.([]byte), &data)
		if err != nil {
			log.Debug("RedisCache: Error deserialize data")
			return nil, errors.New("RedisCache: Error deserialize data" + err.Error())
		}
		return data, nil
	}

	log.Debug("RedisCache: Could not find index. Looking up source")
	data, err = source(index)
	if err != nil {
		log.Debug("RedisCache: error looking up source" + err.Error())
		return nil, errors.New("RedisCache: error looking up source" + err.Error())
	}

	if data == nil {
		log.Debug("RedisCache: could not find index in source")
		return nil, errors.New("RedisCache: could not find index in source")
	}
	serializeddata, err := cache.serialize(&data)
	if err != nil {
		log.Debug("RedisCache: Error serializing data")
		return nil, errors.New("RedisCache: Error serializing data" + err.Error())
	}
	reply, err = cache.client.Do("SETEX", cache.keyPrefix.AppendKey(key).Append(index), cache.maxAge, serializeddata)
	if err != nil {
		log.Debug("RedisCache: Error saving index")
		return nil, errors.New("RedisCache: Error saving index" + err.Error())
	}

	return data, nil
}

func (cache *RedisCache) serialize(data *interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (cache *RedisCache) deserialize(serializeddata []byte, data *interface{}) error {
	buffer := bytes.NewBuffer(serializeddata)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&data)
}

func (cache *RedisCache) Lock() {
	cache.mux.Lock()
}

func (cache *RedisCache) Unlock() {
	cache.mux.Unlock()
}
