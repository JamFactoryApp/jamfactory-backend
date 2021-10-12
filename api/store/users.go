package store

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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

func (s *RedisUserStore) New(identifier string, username string, usertype users.UserType, token *oauth2.Token) *users.User {
	return &users.User{
		Identifier:   identifier,
		UserType:     usertype,
		UserName:     username,
		SpotifyToken: token,
	}
}

func (s *RedisUserStore) NewEmpty() *users.User {
	return &users.User{
		Identifier:   "",
		UserType:     users.UserTypeEmpty,
		UserName:     "",
		SpotifyToken: nil,
	}
}

func (s *RedisUserStore) Get(identifier string) (*users.User, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("GET", s.redisKey.Append(identifier))
	var user users.User
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, errors.New("RedisUserStore: user not found")
	}
	if data, ok := reply.([]byte); ok {
		err = s.deserialize(data, &user)
	} else {
		err = errors.New("RedisUserStore: Failed to convert user from interface{} to []bytes")
	}

	return &user, err
}

func (s *RedisUserStore) Save(user *users.User) error {
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

func (s *RedisUserStore) serialize(user *users.User) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(user)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *RedisUserStore) deserialize(data []byte, user *users.User) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(user)
}
