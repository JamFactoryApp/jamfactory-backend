package jamfactory

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/server"
	"github.com/jamfactoryapp/jamfactory-backend/internal/logutils"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamlabel"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

type SpotifyJamFactory struct {
	authenticator spotify.Authenticator
	cache         cache.Cache
	labelManager  jamlabel.Manager
	jamSessions   map[string]jamsession.JamSession
	clientAddress string
	log           *log.Logger
}

var (
	scopes = []string{
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
	}
)

func NewSpotify(ca *cache.RedisCache, redirectURL string, clientID string, secretKey string, clientAddress string) server.JamFactory {
	a := spotify.NewAuthenticator(redirectURL, scopes...)
	a.SetAuthInfo(clientID, secretKey)
	return &SpotifyJamFactory{
		authenticator: a,
		cache:         ca,
		labelManager:  jamlabel.NewDefault(),
		jamSessions:   make(map[string]jamsession.JamSession),
		clientAddress: clientAddress,
		log:           logutils.NewDefault(),
	}
}

func (s *SpotifyJamFactory) Authenticate(state string, r *http.Request) (*oauth2.Token, error) {
	return s.authenticator.Token(state, r)
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

	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Close,
		Message: notifications.HostLeft,
	})

	if err := jamSession.Delete(); err != nil {
		return err
	}

	return nil
}

func (s *SpotifyJamFactory) GetJamSession(jamLabel string) (jamsession.JamSession, error) {
	jamSession, exists := s.jamSessions[strings.ToUpper(jamLabel)]
	if !exists {
		return nil, apierrors.ErrJamSessionNotFound
	}
	return jamSession, nil
}

func (s *SpotifyJamFactory) NewJamSession(token *oauth2.Token) (jamsession.JamSession, error) {
	client := s.authenticator.NewClient(token)
	jamSession, err := jamsession.NewSpotify(client, s.labelManager.Create())
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
