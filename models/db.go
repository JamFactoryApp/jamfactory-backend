package models

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
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
			address := os.Getenv("JAM_REDIS_ADDRESS")
			port, err := strconv.Atoi(os.Getenv("JAM_REDIS_PORT"))
			if err != nil {
				log.Fatal("Invalid redis port")
			}
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
			if err != nil {
				return nil, err
			}
			if password, ok := os.LookupEnv("JAM_REDIS_PASSWORD"); ok && password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					_ = c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", os.Getenv("JAM_REDIS_DATABASE")); err != nil {
				_ = c.Close()
				return nil, err
			}
			return c, err
		},
	}
	return pool
}
