package models

import (
	"github.com/gomodule/redigo/redis"
	"os"
	"time"
)

var (
	RedisPool *redis.Pool
)

func initRedisPool() {
	RedisPool = newRedisPool()
}

func newRedisPool() *redis.Pool {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", os.Getenv("REDIS_ADDRESS"))
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", os.Getenv("REDIS_DATABASE")); err != nil {
				_ = c.Close()
				return nil, err
			}
			return c, err
		},
	}
	return pool
}
