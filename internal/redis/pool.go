package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

const maxIdle = 3
const idleTimeout = 240 * time.Second

func NewPool(address string, password string, database string) (*redis.Pool, error) {
	pool := &redis.Pool{
		MaxIdle:     maxIdle,
		IdleTimeout: idleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s", address))
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					_ = c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", database); err != nil {
				_ = c.Close()
				return nil, err
			}
			return c, err
		},
	}
	conn := pool.Get()
	_, err := conn.Do("PING")
	if err != nil {
		return nil, err
	}
	return pool, nil
}
