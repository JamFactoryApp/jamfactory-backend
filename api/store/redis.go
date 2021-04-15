package store

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"encoding/gob"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/internal/ioutils"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	cookieKeyLength            = 32
	cookieMaxAge               = 60 * 60 * 24 * 7 // Cookie can last for 7 days
	sessionMaxAge              = 60 * 60 * 24 * 2 // Session can last for 2 days
	defaultRedisSessionKey     = "session"
	defaultCookieKeyPairsCount = 4
	minCookieKeyPairsCount     = 4
)

type RedisStore struct {
	sync.Mutex
	pool          *redis.Pool
	redisKey      pkgredis.Key
	options       *sessions.Options
	codecs        []securecookie.Codec
	keyPairsCount int
	keyPairsFile  string
}

func NewRedis(pool *redis.Pool, keyPairsFile string, sameSite http.SameSite, secure bool) *RedisStore {
	redisStore := &RedisStore{
		pool:     pool,
		redisKey: pkgredis.Key{}.Append(defaultRedisSessionKey),
		options: &sessions.Options{
			Path:     "/",
			MaxAge:   cookieMaxAge,
			SameSite: sameSite,
			Secure:   secure,
		},
		keyPairsCount: defaultCookieKeyPairsCount,
		keyPairsFile:  keyPairsFile,
	}

	redisStore.MaxAge(redisStore.options.MaxAge)
	redisStore.LoadCookieKeyPairs()
	return redisStore
}

func (s *RedisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.options
	session.Options = &opts
	session.IsNew = true

	var err error
	if cookie, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, cookie.Value, &session.ID, s.codecs...)
		if err == nil {
			ok, err := s.load(session)
			session.IsNew = !(err == nil && ok)
		}
	}

	return session, err
}

func (s *RedisStore) Save(_ *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}

		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = s.generateID()
	}

	if err := s.save(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *RedisStore) MaxAge(age int) {
	s.options.MaxAge = age

	for _, codec := range s.codecs {
		if secureCookie, ok := codec.(*securecookie.SecureCookie); ok {
			secureCookie.MaxAge(age)
		}
	}
}

func (s *RedisStore) LoadCookieKeyPairs() {
	var keyPairs [][]byte
	if ioutils.FileExists(s.keyPairsFile) {
		keyPairs = s.readCookieKeyPairs()
	} else {
		keyPairs = s.generateCookieKeyPairs()
		s.writeCookieKeyPairs(keyPairs)
	}
	s.codecs = securecookie.CodecsFromPairs(keyPairs...)
}

func (s *RedisStore) load(session *sessions.Session) (bool, error) {
	s.Lock()
	defer s.Unlock()

	conn := s.pool.Get()
	reply, err := conn.Do("GET", s.redisKey.Append(session.ID))
	if err != nil {
		return false, err
	}
	if reply == nil {
		return false, errors.New("RedisStore: session not found")
	}
	if data, ok := reply.([]byte); ok {
		err = s.deserialize(data, session)
	} else {
		err = errors.New("RedisStore: Failed to convert session data from interface{} to []bytes")
	}

	if reply, err = conn.Do("EXPIRE", s.redisKey.Append(session.ID), sessionMaxAge); err != nil {
		log.Error("RedisStore: Failed to update expiry")
	}

	return true, err
}

func (s *RedisStore) save(session *sessions.Session) error {
	conn := s.pool.Get()
	serialized, err := s.serialize(session)
	if err != nil {
		return err
	}
	reply, err := conn.Do("SET", s.redisKey.Append(session.ID), serialized, "EX", sessionMaxAge)
	log.Trace("redis reply (DO SET): ", reply, " with err: ", err)
	return err
}

func (s *RedisStore) delete(session *sessions.Session) error {
	conn := s.pool.Get()
	_, err := conn.Do("DEL", s.redisKey.Append(session.ID))
	return err
}

func (s *RedisStore) readCookieKeyPairs() [][]byte {
	file, err := os.Open(s.keyPairsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer ioutils.CloseProperly(file)

	r := bufio.NewReader(file)

	keyPairs := make([][]byte, 2*s.keyPairsCount)
	for i := 0; i < s.keyPairsCount*2; i++ {
		keyPairs[i] = make([]byte, cookieKeyLength)
		n, err := io.ReadFull(r, keyPairs[i])
		if n != cookieKeyLength {
			log.Fatalf("Error parsing %s\n", s.keyPairsFile)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(keyPairs) != s.keyPairsCount*2 {
		log.Fatalf("wrong number of cookie key pairs in %s\n", s.keyPairsFile)
	}

	return keyPairs
}

func (s *RedisStore) writeCookieKeyPairs(keyPairs [][]byte) {
	file, err := os.Create(s.keyPairsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer ioutils.CloseProperly(file)

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

func (s *RedisStore) generateCookieKeyPairs() [][]byte {
	var count int
	var keyPairs [][]byte

	if s.keyPairsCount < minCookieKeyPairsCount {
		count = minCookieKeyPairsCount
	} else {
		count = s.keyPairsCount
	}
	keyPairs = make([][]byte, 2*count)

	for i := range keyPairs {
		keyPairs[i] = securecookie.GenerateRandomKey(cookieKeyLength)
	}

	return keyPairs
}

func (s *RedisStore) serialize(session *sessions.Session) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(session.Values)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (s *RedisStore) deserialize(data []byte, session *sessions.Session) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&session.Values)
}

func (s *RedisStore) generateID() string {
	return strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
}
