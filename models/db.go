package models

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	RedisAddress = os.Getenv("REDIS_ADDRESS")
)

var (
	rdb *redis.Client
)

func initRedisClient() {
	rdb = redis.NewClient(&redis.Options{
		Addr: RedisAddress,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		log.Panic("Error connecting to redis: ", err.Error())
	}
}
