package jamsession

import (
	"bytes"
	"encoding/gob"

	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
)

const (
	defaultRedisJamKey = "key"
)

type Store struct {
	pool     *redis.Pool
	redisKey pkgredis.Key
}

func NewRedisJamStore(pool *redis.Pool) *Store {
	return &Store{
		pool:     pool,
		redisKey: pkgredis.Key{}.Append(defaultRedisJamKey),
	}
}

func (s *Store) Get(label string) (*JamSession, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("GET", s.redisKey.Append(label))
	var jam *JamSession
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrJamNotFound
	}
	if data, ok := reply.([]byte); ok {
		err = s.deserialize(data, jam)
	} else {
		err = ErrInterfaceConvert
	}

	return jam, err
}

func (s *Store) GetAll() ([]*JamSession, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("SCAN", s.redisKey.Append("*"))
	var jams []*JamSession
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ErrJamNotFound
	}

	log.Info(reply)

	return jams, err
}

func (s *Store) Save(jam *JamSession) error {
	conn := s.pool.Get()
	serialized, err := s.serialize(jam)
	if err != nil {
		return err
	}
	reply, err := conn.Do("SET", s.redisKey.Append(jam.JamLabel()), serialized)
	log.Trace("redis reply (DO SET): ", reply, " with err: ", err)
	return err
}

func (s *Store) Delete(identifier string) error {
	conn := s.pool.Get()
	_, err := conn.Do("DEL", s.redisKey.Append(identifier))
	return err
}

func (s *Store) serialize(jam *JamSession) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(jam)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *Store) deserialize(data []byte, jam *JamSession) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(jam)
}
