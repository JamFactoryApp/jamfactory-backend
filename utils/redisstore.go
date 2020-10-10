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
	"io"
	"log"
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
	CookieKeyPairsFile     = "/go/data/.keypairs"
)

type RedisStore struct {
	client        redis.Conn
	keyPrefix     RedisKey
	options       *sessions.Options
	codecs        []securecookie.Codec
	keyPairsCount int
	mux           sync.Mutex
}

func NewRedisStore(client redis.Conn, keyPrefix RedisKey, maxAge int, keyPairsCount int) *RedisStore {
	redisStore := &RedisStore{
		client:    client,
		keyPrefix: keyPrefix,
		options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			Secure:   true,
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
			err = store.load(session)
			if err == nil {
				session.IsNew = false
			}
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

func (store *RedisStore) load(session *sessions.Session) error {
	store.mux.Lock()
	reply, err := store.client.Do("GET", store.keyPrefix.Append(session.ID))
	store.mux.Unlock()
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

func (store *RedisStore) save(session *sessions.Session) error {
	serialized, err := store.serializeSession(session)
	if err != nil {
		return err
	}
	_, err = store.client.Do("SET", store.keyPrefix.Append(session.ID), serialized)
	return err
}

func (store *RedisStore) delete(session *sessions.Session) error {
	_, err := store.client.Do("DEL", store.keyPrefix.Append(session.ID))
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
