package hub

import (
	"context"
	"errors"

	"github.com/jamfactoryapp/jamfactory-backend/pkg/authenticator"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/store"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Hub struct {
	Authenticator *authenticator.Authenticator
	Stores
	users map[string]*users.User
}

type Stores struct {
	Store       store.Store[users.UserInformation]
	Identifiers store.Set
}

func NewHub(authenticator *authenticator.Authenticator, stores Stores) *Hub {
	hub := &Hub{
		Authenticator: authenticator,
		Stores:        stores,
		users:         make(map[string]*users.User),
	}
	return hub
}

func (h *Hub) NewUser(ctx context.Context, id string, username string, userType users.UserType, token *oauth2.Token) (*users.User, error) {
	user, err := users.New(ctx, id, username, userType, h.Store, token, h.Authenticator)
	if err != nil {
		return nil, err
	}
	err = h.Identifiers.Add(id)
	if err != nil {
		return nil, err
	}
	h.users[id] = user
	return user, nil
}

func (h *Hub) GetUserByIdentifier(ctx context.Context, identifier string) (*users.User, error) {
	// Check if local user exists
	user, ok := h.users[identifier]
	if ok {
		return user, nil
	}
	log.Trace("User not found local")

	// Check if user identifier exists in store
	exists, err := h.Identifiers.Has(identifier)
	if err != nil {
		return nil, err
	}

	if exists {
		log.Trace("User found in store")
		user = users.Load(ctx, identifier, h.Store, h.Authenticator)
		h.users[identifier] = user
		if err != nil {
			return nil, err
		}

		return user, nil
	}
	log.Trace("User not found")
	return nil, ErrUserNotFound

}

func (h *Hub) DeleteUser(ctx context.Context, identifier string) {
	_, err := h.GetUserByIdentifier(ctx, identifier)
	if err != nil {

	}

	if err := h.Stores.Identifiers.Delete(identifier); err != nil {

	}
	if err := h.Stores.Store.Delete(identifier); err != nil {

	}

	delete(h.users, identifier)

}
