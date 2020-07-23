package models

import (
	"encoding/gob"
	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

const (
	storeMaxAge = 3600

	SessionUserTypeKey  = "User"
	SessionLabelTypeKey = "Label"
	SessionTokenKey     = "Token"
)

var (
	Store *redisstore.RedisStore
)

type Session struct {
	ID string
	//Data string
	Data     map[string]interface{}
	Modified time.Time
}

type CookieToken struct{}

type TokenGetterSetter interface {
	GetToken(r *http.Request, name string) (string, error)
	SetToken(w http.ResponseWriter, name string, value string, options *sessions.Options)
}

func initSessionStore() {
	gob.Register(oauth2.Token{})

	var err error
	Store, err = redisstore.NewRedisStore(rdb)
	if err != nil {
		log.Fatal("Failed to create redis store: ", err)
	}

	Store.KeyPrefix("")
	Store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   storeMaxAge,
		SameSite: http.SameSiteNoneMode,
		Secure:   false,
	})
}

func (token *CookieToken) GetToken(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func (token *CookieToken) SetToken(r http.ResponseWriter, name string, value string, options *sessions.Options) {
	http.SetCookie(r, sessions.NewCookie(name, value, options))
}
