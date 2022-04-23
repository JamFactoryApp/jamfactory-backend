package jamsession

import (
	"bytes"
	"encoding/gob"
	"github.com/gomodule/redigo/redis"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	log "github.com/sirupsen/logrus"
	"strconv"
)

const (
	defaultRedisJamKey = "jam"
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
	jam, err := s.get(s.redisKey.Append(label).String())
	return jam, err
}

func (s *Store) get(key string) (*JamSession, error) {
	conn := s.pool.Get()
	reply, err := conn.Do("GET", key)
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
	jams := make([]*JamSession, len(keys))
	i := 0
	for key, _ := range keys {
		jam, err := s.get(key)
		if err != nil {
			return nil, err
		}
		jams[i] = jam
		i++
	}
	log.Info(jams)

	return jams, nil
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
