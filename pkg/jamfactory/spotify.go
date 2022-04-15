package jamfactory

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	"strings"
	"time"

	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/internal/logutils"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

type JamFactory struct {
	cache *cache.Cache
	store *jamsession.Store
	users *users.Store
	log   *log.Logger
}

const (
	inactiveTime    = 2 * time.Hour
	inactiveWarning = 1*time.Hour + 30*time.Minute
)

func New(store *jamsession.Store, users *users.Store, ca *cache.Cache) *JamFactory {
	jamFactory := &JamFactory{
		cache: ca,
		store: store,
		users: users,
		log:   logutils.NewDefault(),
	}
	go jamFactory.Housekeeper()
	return jamFactory
}

func (s *JamFactory) Housekeeper() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		jamSessions, err := s.store.GetAll()
		if err != nil {
			log.Warn(err)
		}
		for _, jamSession := range jamSessions {
			if time.Now().After(jamSession.Timestamp().Add(inactiveWarning)) {
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Warning,
				})
			}

			if time.Now().After(jamSession.Timestamp().Add(inactiveTime)) {
				log.Debug(jamSession.JamLabel(), ": inactive, closing")
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Inactive,
				})
				if err := s.DeleteJamSession(jamSession.JamLabel()); err != nil {
					s.log.Debug(err)
				}
			}
		}
	}
}

func (s *JamFactory) DeleteJamSession(jamLabel string) error {
	jamSession, err := s.store.Get(jamLabel)
	if err != nil {
		return apierrors.ErrJamSessionNotFound
	}

	if err := jamSession.Deconstruct(); err != nil {
		return err
	}

	if err := s.store.Delete(jamLabel); err != nil {
		return err
	}

	return nil
}

func (s *JamFactory) GetJamSessionByLabel(jamLabel string) (*jamsession.JamSession, error) {
	jamSession, err := s.store.Get(strings.ToUpper(jamLabel))
	if err != nil {
		return nil, apierrors.ErrJamSessionNotFound
	}
	return jamSession, nil
}

func (s *JamFactory) GetJamSessionByUser(user *users.User) (*jamsession.JamSession, error) {
	jamSessions, err := s.store.GetAll()
	if err != nil {
		log.Warn(err)
	}
	for _, jamSession := range jamSessions {
		if _, err := jamSession.Members().Get(user.Identifier); err == nil {
			return jamSession, nil
		}
	}
	return nil, apierrors.ErrJamSessionNotFound
}

func (s *JamFactory) NewJamSession(host *users.User) (*jamsession.JamSession, error) {
	// Check if correct user type was passed
	if host.UserType != users.UserTypeSpotify {
		return nil, errors.New("Wrong userIdentifier Type for Spotify JamSession with UserType: " + string(host.UserType))
	}

	jamSession, err := jamsession.New(host, s.users, s.CreateLabel(0))
	if err != nil {
		return nil, err
	}
	err = s.store.Save(jamSession)
	if err != nil {
		return nil, err
	}
	return jamSession, nil
}

func (s *JamFactory) Search(jamSession *jamsession.JamSession, searchType string, text string) (interface{}, error) {
	country := spotify.CountryGermany
	opts := spotify.Options{
		Country: &country,
	}

	var spotifySearchType spotify.SearchType
	var key = pkgredis.NewKey("search")
	switch searchType {
	case "track":
		spotifySearchType = spotify.SearchTypeTrack
		key = key.Append(searchType)
	case "playlist":
		spotifySearchType = spotify.SearchTypePlaylist
		key = key.Append(searchType)
	case "album":
		spotifySearchType = spotify.SearchTypeAlbum
		key = key.Append(searchType)
	}
	if spotifySearchType == 0 {
		return nil, apierrors.ErrSearchTypeInvalid
	}

	searchString := []string{text, "*"}
	entry, err := s.cache.Query(key, strings.Join(searchString, ""), func(index string) (interface{}, error) {
		return jamSession.Search(index, spotifySearchType, &opts)
	})

	if err != nil {
		return nil, err
	}

	result, ok := entry.(*spotify.SearchResult)
	if !ok {
		return nil, apierrors.ErrSearchResultMalformed
	}

	return result, nil
}
