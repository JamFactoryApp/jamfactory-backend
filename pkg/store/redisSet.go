package store

import (
	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
)

type RedisSet struct {
	pool     *redis.Pool
	redisKey pkgredis.Key
}

func NewRedisSet(pool *redis.Pool, key string) *RedisSet {
	return &RedisSet{
		pool:     pool,
		redisKey: pkgredis.Key{}.Append(key),
	}
}

func (s RedisSet) GetAll() ([]string, error) {
	conn := s.pool.Get()
	reply, err := redis.Strings(conn.Do("SMEMBERS", s.redisKey))
	if err != nil {
		return nil, err
	}
	log.Trace("redis DO SMEMBERS", " with err: ", err)
	return reply, nil
}

func (s RedisSet) Add(key string) error {
	conn := s.pool.Get()
	_, err := conn.Do("SADD", s.redisKey, key)
	log.Trace("redis DO SADD for: ", key, " with err: ", err)
	return err
}

func (s RedisSet) Has(key string) (bool, error) {
	conn := s.pool.Get()
	reply, err := redis.Int(conn.Do("SISMEMBER", s.redisKey, key))
	log.Trace("redis DO SISMEMBER for: ", key, " result: ", reply == 1, " with err: ", err)
	return reply == 1, nil
}

func (s RedisSet) Delete(key string) error {
	conn := s.pool.Get()
	_, err := conn.Do("SREM", s.redisKey, key)
	log.Trace("redis DO SREM for: ", key, " with err: ", err)
	return err
}
