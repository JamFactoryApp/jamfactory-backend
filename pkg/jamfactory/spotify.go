package jamfactory

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/hub"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/queue"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/store"
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

type Stores struct {
	JamLabels store.Set
	Settings  store.Store[jamsession.Settings]
	Members   store.Store[jamsession.Members]
	Queues    store.Store[queue.Queue]
}

type JamFactory struct {
	JamSessions map[string]*jamsession.JamSession
	hub         *hub.Hub
	cache       *cache.Cache
	log         *log.Logger
	Stores
}

const (
	inactiveTime    = 2 * time.Hour
	inactiveWarning = 1*time.Hour + 30*time.Minute
)

func New(stores Stores, hub *hub.Hub, ca *cache.Cache) *JamFactory {
	jamFactory := &JamFactory{
		JamSessions: make(map[string]*jamsession.JamSession),
		cache:       ca,
		Stores:      stores,
		hub:         hub,
		log:         logutils.NewDefault(),
	}
	go jamFactory.Housekeeper()
	return jamFactory
}

func (s *JamFactory) Housekeeper() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C

		for _, jamSession := range s.JamSessions {
			if time.Now().After(jamSession.Timestamp.Add(inactiveWarning)) {
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Warning,
				})
			}

			if time.Now().After(jamSession.Timestamp.Add(inactiveTime)) {
				log.Debug(jamSession.JamLabel, ": inactive, closing")
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Inactive,
				})
				if err := s.DeleteJamSession(jamSession.JamLabel); err != nil {
					s.log.Debug(err)
				}
			}
		}
	}
}

func (s *JamFactory) DeleteJamSession(jamLabel string) error {
	jamSession, err := s.GetJamSessionByLabel(jamLabel)
	if err != nil {
		return apierrors.ErrJamSessionNotFound
	}

	if err := jamSession.Deconstruct(); err != nil {
		return err
	}
	if err := s.Stores.Members.Delete(jamLabel); err != nil {
		return err
	}
	if err := s.Stores.Settings.Delete(jamLabel); err != nil {
		return err
	}
	if err := s.Stores.Queues.Delete(jamLabel); err != nil {
		return err
	}

	delete(s.JamSessions, jamLabel)

	return nil
}

func (s *JamFactory) GetJamSessionByLabel(jamLabel string) (*jamsession.JamSession, error) {
	// Check if local JamSession exists
	jamSession, ok := s.JamSessions[jamLabel]
	if ok {
		return jamSession, nil
	}
	log.Trace("JamSession not found local")

	// Check if label exists in store
	exists, err := s.JamLabels.Has(jamLabel)
	if err != nil {
		return nil, err
	}

	if exists {
		log.Trace("JamSession found in store")
		stores := jamsession.Stores{
			Members:  s.Members,
			Queues:   s.Queues,
			Settings: s.Settings,
		}
		jamSession, err = jamsession.Load(stores, s.hub, jamLabel)
		s.JamSessions[jamLabel] = jamSession
		if err != nil {
			return nil, err
		}

		return jamSession, nil
	}
	log.Trace("JamSession not found")
	return nil, jamsession.ErrJamSessionMissing

}

func (s *JamFactory) GetJamSessionByUser(user *users.User) (*jamsession.JamSession, error) {
	jamLabels, err := s.JamLabels.GetAll()
	if err != nil {
		log.Warn(err)
	}
	log.Warn(jamLabels)
	for _, jamLabel := range jamLabels {
		log.Warn(jamLabel)
		jamSession, err := s.GetJamSessionByLabel(jamLabel)
		if err != nil {
			log.Warn(err)
			continue
		}
		members := jamSession.GetMembers()
		log.Warn(members)
		if _, err := members.Get(user.Identifier); err == nil {
			return jamSession, nil
		}
	}
	return nil, apierrors.ErrJamSessionNotFound
}

func (s *JamFactory) NewJamSession(host *users.User) (*jamsession.JamSession, error) {
	// Check if correct user type was passed
	if host.GetInfo().UserType != users.UserTypeSpotify {
		return nil, errors.New("Wrong userIdentifier Type for Spotify JamSession with UserType: " + string(host.GetInfo().UserType))
	}

	stores := jamsession.Stores{
		Members:  s.Members,
		Queues:   s.Queues,
		Settings: s.Settings,
	}

	jamLabel := s.CreateLabel(0)

	jamSession, err := jamsession.CreateNew(host, stores, s.hub, jamLabel)
	if err != nil {
		return nil, err
	}
	err = s.JamLabels.Add(jamLabel)
	if err != nil {
		return nil, err
	}
	s.JamSessions[jamLabel] = jamSession
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
