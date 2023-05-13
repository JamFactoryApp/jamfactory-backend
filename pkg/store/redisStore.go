package store

import (
	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type RedisStore[T any] struct {
	pool     *redis.Pool
	redisKey pkgredis.Key
}

func NewRedisStore[T any](pool *redis.Pool, key string) *RedisStore[T] {
	return &RedisStore[T]{
		pool:     pool,
		redisKey: pkgredis.Key{}.Append(key),
	}
}

func (s RedisStore[T]) Get(key string) (*T, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("GET", s.redisKey.Append(key).String())
	log.Trace("redis DO GET for: ", key, " with err: ", err)
	obj := new(T)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrObjNotFound
	}
	if data, ok := reply.([]byte); ok {
		err = deserialize(data, obj)
	} else {
		return nil, ErrInterfaceConvert
	}

	return obj, err
}

func (s RedisStore[T]) GetAll() ([]*T, error) {
	conn := s.pool.Get()
	cursor := 0
	maxRecursion := 100
	keys := make(map[string]bool, 0)
	for {
		maxRecursion--
		reply, err := conn.Do("SCAN", cursor, "MATCH", s.redisKey.Append("*"))
		if err != nil {
			return nil, err
		}
		replyArr := reply.([]interface{})
		cursorString := string(replyArr[0].([]uint8))
		cursor, err = strconv.Atoi(cursorString)
		if err != nil {
			return nil, err
		}

		for _, key := range replyArr[1].([]interface{}) {
			keys[string(key.([]byte))] = true
		}

		if cursor == 0 || maxRecursion == 0 {
			break
		}
	}
	jams := make([]*T, len(keys))
	i := 0
	for key := range keys {
		jam, err := s.Get(key)
		if err != nil {
			return nil, err
		}
		jams[i] = jam
		i++
	}
	return jams, nil
}

func (s RedisStore[T]) Save(obj *T, key string) error {
	conn := s.pool.Get()
	serialized, err := serialize(obj)
	if err != nil {
		return err
	}
	_, err = conn.Do("SET", s.redisKey.Append(key), serialized)
	log.Trace("redis DO SET for: ", key, " with err: ", err)
	return err
}

func (s RedisStore[T]) Delete(key string) error {
	conn := s.pool.Get()
	_, err := conn.Do("DEL", s.redisKey.Append(key))
	log.Trace("redis DO DEL for: ", key, " with err: ", err)
	return err
}
