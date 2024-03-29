package cache

import (
	"bytes"
	"encoding/gob"
	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/pkg/errors"
	"sync"
)

var (
	ErrDeserializeData       = errors.New("failed to deserialize data")
	ErrIndexNotFoundInSource = errors.New("could not find index in source")
	ErrLookupSource          = errors.New("failed to lookup source")
	ErrRedisQuery            = errors.New("failed to query redis")
	ErrSaveIndex             = errors.New("failed to save index")
	ErrSerializeData         = errors.New("failed to serialize data")
)

const (
	defaultKeyPrefix = "cache"
	defaultTTL       = 60 * 60 * 24 // Cache entries should update every day
)

type Cache struct {
	sync.Mutex
	pool      *redis.Pool
	keyPrefix pkgredis.Key
	ttl       int
}

func NewRedis(pool *redis.Pool) *Cache {
	redisCache := &Cache{
		pool:      pool,
		keyPrefix: pkgredis.Key{}.Append(defaultKeyPrefix),
		ttl:       defaultTTL,
	}

	return redisCache
}

func (c *Cache) Query(key pkgredis.Key, index string, source SourceFunc) (interface{}, error) {
	c.Lock()
	defer c.Unlock()

	conn := c.pool.Get()

	reply, err := conn.Do("GET", c.keyPrefix.AppendKey(key).Append(index))

	var data interface{}
	if err != nil {
		return nil, errors.Wrap(err, ErrRedisQuery.Error())
	}

	if reply != nil {
		err = c.deserialize(reply.([]byte), &data)
		if err != nil {
			return nil, errors.Wrap(err, ErrDeserializeData.Error())
		}
		return data, nil
	}

	data, err = source(index)
	if err != nil {
		return nil, errors.Wrap(err, ErrLookupSource.Error())
	}

	if data == nil {
		return nil, errors.Wrap(err, ErrIndexNotFoundInSource.Error())
	}
	serializeddata, err := c.serialize(&data)
	if err != nil {
		return nil, errors.Wrap(err, ErrSerializeData.Error())
	}
	reply, err = conn.Do("SETEX", c.keyPrefix.AppendKey(key).Append(index), c.ttl, serializeddata)
	if err != nil {
		return nil, errors.Wrap(err, ErrSaveIndex.Error())
	}

	return data, nil
}

func (c *Cache) serialize(data *interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c *Cache) deserialize(serializeddata []byte, data *interface{}) error {
	buffer := bytes.NewBuffer(serializeddata)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&data)
}
