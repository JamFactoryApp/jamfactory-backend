package models

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"strings"
)

const (
	SessionUserTypeKey  = "User"
	SessionLabelTypeKey = "Label"
	SessionTokenKey     = "Token"
)

type RedisStore struct {
	client    redis.Conn
	keyPrefix string
	options   *sessions.Options
	codecs    []securecookie.Codec
}

func NewRedisStore(client redis.Conn, keyPrefix string, maxAge int, keyPairs ...[]byte) *RedisStore {
	redisStore := &RedisStore{
		client:    client,
		keyPrefix: keyPrefix,
		options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		},
		codecs: securecookie.CodecsFromPairs(keyPairs...),
	}
	redisStore.MaxAge(redisStore.options.MaxAge)
	return redisStore
}

func (store *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(store, name)
}

func (store *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(store, name)
	opts := *store.options
	session.Options = &opts
	session.IsNew = true

	cookie, errCookie := r.Cookie(name)
	if errCookie != nil {
		return session, nil
	}
	//err := securecookie.DecodeMulti(name, cookie.Value, &session.ID, store.codecs...)
	session.ID = cookie.Value

	err := store.load(session)
	if err == nil {
		session.IsNew = false
	} else if err == redis.ErrNil {
		err = nil
	}

	return session, err
}

func (store *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		if err := store.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		id, err := store.idGen()
		if err != nil {
			return errors.New("RedisStore: Failed to generate session id")
		}
		session.ID = id
	}

	//encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.codecs...)
	//if err != nil {
	//	return err
	//}

	if err := store.save(session); err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), session.ID, session.Options))
	return nil
}

func (store RedisStore) MaxAge(age int) {
	store.options.MaxAge = age

	for _, codec := range store.codecs {
		if secureCookie, ok := codec.(*securecookie.SecureCookie); ok {
			secureCookie.MaxAge(age)
		}
	}
}

func (store RedisStore) load(session *sessions.Session) error {
	reply, err := store.client.Do("GET", store.keyPrefix+session.ID)
	if err != nil {
		return err
	}
	if reply == nil {
		return errors.New("RedisStore: session not found")
	}
	if data, ok := reply.([]byte); ok {
		err = store.deserializeSession(data, session)
	} else {
		err = errors.New("RedisStore: Failed to convert session data from interface{} to []bytes")
	}
	return err
}

func (store RedisStore) save(session *sessions.Session) error {
	serialized, err := store.serializeSession(session)
	if err != nil {
		return err
	}
	_, err = store.client.Do("SET", store.keyPrefix+session.ID, serialized)
	return err
}

func (store RedisStore) delete(session *sessions.Session) error {
	_, err := store.client.Do("DELETE", store.keyPrefix+session.ID)
	return err
}

func (store RedisStore) serializeSession(session *sessions.Session) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(session.Values)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (store RedisStore) deserializeSession(data []byte, session *sessions.Session) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&session.Values)
}

func (store RedisStore) idGen() (string, error) {
	key := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.EncodeToString(key), "="), nil
}
