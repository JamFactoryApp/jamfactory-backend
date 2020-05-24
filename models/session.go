package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"time"
)

type Session struct {
	ID primitive.ObjectID `bson:"_id, omitempty"`
	//Data string
	Data     map[string]interface{}
	Modified time.Time
}

type SessionStore struct {
	Codecs     []securecookie.Codec
	Options    *sessions.Options
	Token      TokenGetterSetter
	collection *mongo.Collection
}

type CookieToken struct{}

type TokenGetterSetter interface {
	GetToken(r *http.Request, name string) (string, error)
	SetToken(w http.ResponseWriter, name string, value string, options *sessions.Options)
}

func NewSessionStore(collection *mongo.Collection, maxAge int, keyPairs ...[]byte) *SessionStore {
	store := &SessionStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
			SameSite: http.SameSiteLaxMode,
		},
		Token:      &CookieToken{},
		collection: collection,
	}
	store.MaxAge(maxAge)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	indexOptions := options.IndexOptions{}
	indexOptions.SetName("TTL")
	indexOptions.SetExpireAfterSeconds(int32(maxAge))

	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"modified": 1,
		},
		Options: &indexOptions,
	})

	if err != nil {
		log.Println("Error while indexing session store")
	}

	return store
}

func (store *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(store, name)
}

func (store *SessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(store, name)
	session.Options = &sessions.Options{
		Path:     store.Options.Path,
		MaxAge:   store.Options.MaxAge,
		Domain:   store.Options.Domain,
		Secure:   store.Options.Secure,
		HttpOnly: store.Options.HttpOnly,
	}
	session.IsNew = true
	var err error
	if cookie, errToken := store.Token.GetToken(r, name); errToken == nil {
		err = securecookie.DecodeMulti(name, cookie, &session.ID, store.Codecs...)
		if err == nil {
			err = store.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				err = nil
			}
		}
	}
	return session, err
}

func (store *SessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		if err := store.delete(session); err != nil {
			return err
		}
		store.Token.SetToken(w, session.Name(), "", session.Options)
		return nil
	}

	if session.ID == "" {
		session.ID = primitive.NewObjectID().Hex()
	}

	if err := store.upsert(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.Codecs...)
	if err != nil {
		return err
	}

	store.Token.SetToken(w, session.Name(), encoded, session.Options)
	return nil
}

func (store *SessionStore) MaxAge(age int) {
	store.Options.MaxAge = age

	for _, codec := range store.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (store *SessionStore) load(session *sessions.Session) error {
	s := Session{}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	id, _ := primitive.ObjectIDFromHex(session.ID)

	err := store.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&s)

	if err != nil {
		return err
	}

	// Decode Data of session
	//if err := securecookie.DecodeMulti(session.Name(), s.Data, &session.Values,	m.Codecs...); err != nil {
	//	return err
	//}

	// Set session values
	data := make(map[interface{}]interface{})

	for key, value := range s.Data {
		data[key] = value
	}

	session.Values = data

	return nil
}

func (store *SessionStore) upsert(session *sessions.Session) error {
	var modified time.Time
	if val, ok := session.Values["modified"]; ok {
		modified, ok = val.(time.Time)
		if !ok {
			return errors.New("mongostore: invalid modified value")
		}
	} else {
		modified = time.Now()
	}

	// Encode Data of session before storing it in the DB
	//data, err := securecookie.EncodeMulti(session.Name(), session.Values, m.Codecs...)
	//if err != nil {
	//	return err
	//}

	// Create Object containing the Session values. Use only for dev
	data := make(map[string]interface{})

	for key, value := range session.Values {
		strKey := fmt.Sprintf("%v", key)

		data[strKey] = value
	}

	id, _ := primitive.ObjectIDFromHex(session.ID)
	s := Session{
		ID:       id,
		Data:     data,
		Modified: modified,
	}

	filter := bson.D{{"_id", id}}

	ctxCount, cancelCount := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelCount()

	count, err := store.collection.CountDocuments(ctxCount, filter)
	if err != nil {
		return err
	}

	ctxInsert, cancelInsert := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelInsert()

	if count == 0 {
		_, err = store.collection.InsertOne(ctxInsert, s)
		if err != nil {
			return err
		}
	} else {
		update := bson.D{{"$set", s}}
		_, err = store.collection.UpdateOne(ctxInsert, filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (store *SessionStore) delete(session *sessions.Session) error {
	id, _ := primitive.ObjectIDFromHex(session.ID)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.D{{"_id", id}}
	_, err := store.collection.DeleteOne(ctx, filter)

	return err
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
