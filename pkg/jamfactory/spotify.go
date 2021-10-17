package jamfactory

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/server"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	"github.com/jamfactoryapp/jamfactory-backend/internal/logutils"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamlabel"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type SpotifyJamFactory struct {
	authenticator spotify.Authenticator
	cache         cache.Cache
	labelManager  jamlabel.Manager
	jamSessions   map[string]jamsession.JamSession
	clientAddress string
	log           *log.Logger
}

const (
	inactiveTime    = 2 * time.Hour
	inactiveWarning = 1*time.Hour + 30*time.Minute
)

var (
	scopes = []string{
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
	}
)

func NewSpotify(ca *cache.RedisCache, redirectURL string, clientID string, secretKey string, clientAdress string) server.JamFactory {
	a := spotify.NewAuthenticator(redirectURL, scopes...)
	a.SetAuthInfo(clientID, secretKey)
	spotifyJamFactory := &SpotifyJamFactory{
		authenticator: a,
		cache:         ca,
		labelManager:  jamlabel.NewDefault(),
		jamSessions:   make(map[string]jamsession.JamSession),
		clientAddress: clientAdress,
		log:           logutils.NewDefault(),
	}
	go spotifyJamFactory.Housekeeper()
	return spotifyJamFactory
}

func (s *SpotifyJamFactory) Housekeeper() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		for jamLabel, jamSession := range s.jamSessions {
			if time.Now().After(jamSession.Timestamp().Add(inactiveWarning)) {
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Warning,
				})
			}

			if time.Now().After(jamSession.Timestamp().Add(inactiveTime)) {
				log.Debug(jamLabel, ": inactive, closing")
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.Inactive,
				})
				if err := s.DeleteJamSession(jamLabel); err != nil {
					s.log.Debug(err)
				}
			}
		}
	}
}

func (s *SpotifyJamFactory) Authenticate(state string, r *http.Request) (*oauth2.Token, string, string, error) {
	token, err := s.authenticator.Token(state, r)
	if err != nil {
		return nil, "", "", err
	}
	client := s.authenticator.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, "", "", err
	}
	hash := sha1.Sum([]byte(user.Email))
	return token, hex.EncodeToString(hash[:]), user.DisplayName, nil
}

func (s *SpotifyJamFactory) CallbackURL(state string) string {
	return s.authenticator.AuthURL(state)
}

func (s *SpotifyJamFactory) DeleteJamSession(jamLabel string) error {
	jamSession, exists := s.jamSessions[jamLabel]
	if !exists {
		return apierrors.ErrJamSessionNotFound
	}

	if err := s.labelManager.Delete(jamLabel); err != nil {
		s.log.Debug(err)
	}

	if err := jamSession.Deconstruct(); err != nil {
		return err
	}

	delete(s.jamSessions, jamLabel)

	return nil
}

func (s *SpotifyJamFactory) GetJamSessionByLabel(jamLabel string) (jamsession.JamSession, error) {
	jamSession, exists := s.jamSessions[strings.ToUpper(jamLabel)]
	if !exists {
		return nil, apierrors.ErrJamSessionNotFound
	}
	return jamSession, nil
}

func (s *SpotifyJamFactory) GetJamSessionByUser(user *users.User) (jamsession.JamSession, error) {
	for _, jamSession := range s.jamSessions {
		if _, err := jamSession.Members().Get(user.Identifier); err == nil {
			return jamSession, nil
		}
	}
	return nil, apierrors.ErrJamSessionNotFound
}

func (s *SpotifyJamFactory) NewJamSession(host *users.User) (jamsession.JamSession, error) {
	// Check if correct user type was passed
	if host.UserType != users.UserTypeSpotify {
		return nil, errors.New("Wrong userIdentifier Type for Spotify JamSession with UserType: " + string(host.UserType))
	}
	client := s.authenticator.NewClient(host.SpotifyToken)
	jamSession, err := jamsession.NewSpotify(host, client, s.labelManager.Create())
	if err != nil {
		return nil, err
	}
	s.jamSessions[jamSession.JamLabel()] = jamSession
	return jamSession, nil
}

func (s *SpotifyJamFactory) Search(jamSession jamsession.JamSession, t string, text string) (interface{}, error) {
	country := spotify.CountryGermany
	opts := spotify.Options{
		Country: &country,
	}

	var searchType spotify.SearchType
	switch t {
	case "track":
		searchType = spotify.SearchTypeTrack
	case "playlist":
		searchType = spotify.SearchTypePlaylist
	case "album":
		searchType = spotify.SearchTypeAlbum
	}
	if searchType == 0 {
		return nil, apierrors.ErrSearchTypeInvalid
	}

	searchString := []string{text, "*"}
	entry, err := s.cache.Query(pkgredis.NewKey("search"), strings.Join(searchString, ""), func(index string) (interface{}, error) {
		return jamSession.Search(index, searchType, &opts)
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

func (s *SpotifyJamFactory) ClientAddress() string {
	return s.clientAddress
}
