package models

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
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
	Data map[string]interface{}
	Modified time.Time
}

type Sessionstore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	Token   TokenGetSeter
	coll    *mongo.Collection
}

type TokenGetSeter interface {
	GetToken(req *http.Request, name string) (string, error)
	SetToken(rw http.ResponseWriter, name, value string, options *sessions.Options)
}

type CookieToken struct{}

func NewSessionStore(c *mongo.Collection, maxAge int, keyPairs ...[]byte) *Sessionstore {
	store := &Sessionstore{
		Codecs:  securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   maxAge,
		},
		Token:   &CookieToken{},
		coll:    c,
	}
	store.MaxAge(maxAge)

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	indexOptions := options.IndexOptions{}
	indexOptions.SetName("TTL")
	indexOptions.SetExpireAfterSeconds(int32(maxAge))

	c.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.M{
			"modified": 1,
		},
		Options: &indexOptions,
	})

	return store
}

func (m *Sessionstore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(m, name)
}

func (m *Sessionstore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(m, name)
	session.Options = &sessions.Options{
		Path:     m.Options.Path,
		MaxAge:   m.Options.MaxAge,
		Domain:   m.Options.Domain,
		Secure:   m.Options.Secure,
		HttpOnly: m.Options.HttpOnly,
	}
	session.IsNew = true
	var err error
	if cook, errToken := m.Token.GetToken(r, name); errToken == nil {
		err = securecookie.DecodeMulti(name, cook, &session.ID, m.Codecs...)
		if err == nil {
			err = m.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				err = nil
			}
		}
	}
	return session, err
}

func (m *Sessionstore) Save(r *http.Request, w http.ResponseWriter,	session *sessions.Session) error {

	if session.Options.MaxAge < 0 {
		if err := m.delete(session); err != nil {
			return err
		}
		m.Token.SetToken(w, session.Name(), "", session.Options)
		return nil
	}

	if session.ID == "" {
		session.ID = primitive.NewObjectID().Hex()
	}

	if err := m.upsert(session); err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, m.Codecs...)
	if err != nil {
		return err
	}

	m.Token.SetToken(w, session.Name(), encoded, session.Options)
	return nil
}

func (m *Sessionstore) MaxAge(age int) {
	m.Options.MaxAge = age

	for _, codec := range m.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (m *Sessionstore) load(session *sessions.Session) error {

	s := Session{}

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	id, _ := primitive.ObjectIDFromHex(session.ID)

	err := m.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&s)

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

func (m *Sessionstore) upsert(session *sessions.Session) error {

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

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	count, err := m.coll.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)

	if count == 0 {
		_, err = m.coll.InsertOne(ctx, s)
		if err != nil {
			return err
		}
	} else {
		update := bson.D{{"$set", s}}
		_, err = m.coll.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Sessionstore) delete(session *sessions.Session) error {
	id, _ := primitive.ObjectIDFromHex(session.ID)
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	filter := bson.D{{"_id", id}}
	_, err := m.coll.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	return nil

}

func (c *CookieToken) GetToken(req *http.Request, name string) (string, error) {
	cook, err := req.Cookie(name)
	if err != nil {
		return "", err
	}

	return cook.Value, nil
}

func (c *CookieToken) SetToken(rw http.ResponseWriter, name, value string,	options *sessions.Options) {
	http.SetCookie(rw, sessions.NewCookie(name, value, options))
}