package utils

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"encoding/gob"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	SessionUserTypeKey  = "User"
	SessionLabelTypeKey = "Label"
	SessionTokenKey     = "Token"

	MinCookieKeyPairsCount = 4
	CookieKeyLength        = 32
	CookieKeyPairsFile     = "./.keypairs"
)

type RedisStore struct {
	sync.Mutex
	pool          *redis.Pool
	keyPrefix     RedisKey
	options       *sessions.Options
	codecs        []securecookie.Codec
	keyPairsCount int
}

func NewRedisStore(pool *redis.Pool, keyPrefix RedisKey, maxAge int, keyPairsCount int, secureCookies bool) *RedisStore {
	log.Warn(secureCookies)
	redisStore := &RedisStore{
		pool:      pool,
		keyPrefix: keyPrefix,
		options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			Secure:   secureCookies,
			SameSite: http.SameSiteLaxMode,
		},
		keyPairsCount: keyPairsCount,
	}
	redisStore.MaxAge(redisStore.options.MaxAge)
	redisStore.LoadCookieKeyPairs()
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

	var err error
	if cookie, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, cookie.Value, &session.ID, store.codecs...)
		if err == nil {
			ok, err := store.load(session)
			session.IsNew = !(err == nil && ok)
		}
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
		session.ID = store.idGen()
	}

	if err := store.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (store *RedisStore) MaxAge(age int) {
	store.options.MaxAge = age

	for _, codec := range store.codecs {
		if secureCookie, ok := codec.(*securecookie.SecureCookie); ok {
			secureCookie.MaxAge(age)
		}
	}
}

func (store *RedisStore) LoadCookieKeyPairs() {
	var keyPairs [][]byte
	if FileExists(CookieKeyPairsFile) {
		keyPairs = store.readExistingCookieKeyPairs()
	} else {
		keyPairs = store.generateCookieKeyPairs()
		store.saveCookieKeyPairs(keyPairs)
	}
	store.codecs = securecookie.CodecsFromPairs(keyPairs...)
}

func (store *RedisStore) readExistingCookieKeyPairs() [][]byte {
	file, err := os.Open(CookieKeyPairsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseProperly(file)
	r := bufio.NewReader(file)

	keyPairs := make([][]byte, 2*store.keyPairsCount)
	for i := 0; i < store.keyPairsCount*2; i++ {
		keyPairs[i] = make([]byte, CookieKeyLength)
		n, err := io.ReadFull(r, keyPairs[i])
		if n != CookieKeyLength {
			log.Fatalf("Error parsing %s\n", CookieKeyPairsFile)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(keyPairs) != store.keyPairsCount*2 {
		log.Fatalf("wrong number of cookie key pairs in %s\n", CookieKeyPairsFile)
	}
	return keyPairs
}

func (store *RedisStore) saveCookieKeyPairs(keyPairs [][]byte) {
	file, err := os.Create(CookieKeyPairsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseProperly(file)
	w := bufio.NewWriter(file)

	for _, k := range keyPairs {
		if _, err := w.Write(k); err != nil {
			log.Fatal(err)
		}
	}
	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}
}

func (store *RedisStore) generateCookieKeyPairs() [][]byte {
	var count int
	var keyPairs [][]byte

	if store.keyPairsCount < MinCookieKeyPairsCount {
		count = MinCookieKeyPairsCount
	} else {
		count = store.keyPairsCount
	}
	keyPairs = make([][]byte, 2*count)

	for i := range keyPairs {
		keyPairs[i] = securecookie.GenerateRandomKey(CookieKeyLength)
	}

	return keyPairs
}

func (store *RedisStore) load(session *sessions.Session) (bool, error) {
	store.Lock()
	defer store.Unlock()
	conn := store.pool.Get()
	reply, err := conn.Do("GET", store.keyPrefix.Append(session.ID))
	if err != nil {
		return false, err
	}
	if reply == nil {
		return false, errors.New("RedisStore: session not found")
	}
	if data, ok := reply.([]byte); ok {
		err = store.deserializeSession(data, session)
	} else {
		err = errors.New("RedisStore: Failed to convert session data from interface{} to []bytes")
	}
	return true, err
}

func (store *RedisStore) save(session *sessions.Session) error {
	conn := store.pool.Get()
	serialized, err := store.serializeSession(session)
	if err != nil {
		return err
	}
	reply, err := conn.Do("SET", store.keyPrefix.Append(session.ID), serialized)
	log.Trace("redis reply (DO SET): ", reply, " with err: ", err)
	return err
}

func (store *RedisStore) delete(session *sessions.Session) error {
	conn := store.pool.Get()
	_, err := conn.Do("DEL", store.keyPrefix.Append(session.ID))
	return err
}

func (store *RedisStore) serializeSession(session *sessions.Session) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(session.Values)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (store *RedisStore) deserializeSession(data []byte, session *sessions.Session) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&session.Values)
}

func (store *RedisStore) idGen() string {
	return strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
}
