package users

import (
	"bytes"
	"encoding/gob"

	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
)

const (
	defaultRedisUserKey = "user"
)

type RedisUserStore struct {
	pool     *redis.Pool
	redisKey pkgredis.Key
}

func NewRedisUserStore(pool *redis.Pool) *RedisUserStore {
	return &RedisUserStore{
		pool:     pool,
		redisKey: pkgredis.Key{}.Append(defaultRedisUserKey),
	}
}

func (s *RedisUserStore) Get(identifier string) (*User, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("GET", s.redisKey.Append(identifier))
	var user User
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrUserNotFound
	}
	if data, ok := reply.([]byte); ok {
		err = s.deserialize(data, &user)
	} else {
		err = ErrInterfaceConvert
	}

	return &user, err
}

func (s *RedisUserStore) Save(user *User) error {
	conn := s.pool.Get()
	serialized, err := s.serialize(user)
	if err != nil {
		return err
	}
	reply, err := conn.Do("SET", s.redisKey.Append(user.Identifier), serialized)
	log.Trace("redis reply (DO SET): ", reply, " with err: ", err)
	return err
}

func (s *RedisUserStore) Delete(identifier string) error {
	conn := s.pool.Get()
	_, err := conn.Do("DEL", s.redisKey.Append(identifier))
	return err
}

func (s *RedisUserStore) serialize(user *User) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(user)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *RedisUserStore) deserialize(data []byte, user *User) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(user)
}
