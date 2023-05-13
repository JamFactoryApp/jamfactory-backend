package server

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/authenticator"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/config"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/hub"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamfactory"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/queue"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(&oauth2.Token{})
	gob.Register(&spotify.SearchResult{})
	gob.Register(users.UserType(""))
	gob.Register(users.User{})
	gob.Register(jamsession.Settings{})
	gob.Register(jamsession.Members{})
	gob.Register(queue.Queue{})
}

const (
	readTimeout  = time.Second
	writeTimeout = 5 * time.Second
	idleTimeout  = time.Second
)

type Server struct {
	config        *config.Config
	store         *sessions.Store
	server        *http.Server
	router        *mux.Router
	users         *hub.Hub
	cache         *cache.Cache
	authenticator *authenticator.Authenticator
	jamFactory    *jamfactory.JamFactory
	upgrader      websocket.Upgrader
}

func NewServer(pattern string, config *config.Config, sessionStore *sessions.Store, users *hub.Hub, jamFactory *jamfactory.JamFactory, authenticator *authenticator.Authenticator) *Server {
	// Create Authenticator

	s := &Server{
		config: config,
		server: &http.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		router:        mux.NewRouter(),
		authenticator: authenticator,
		store:         sessionStore,
		users:         users,
		jamFactory:    jamFactory,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				log.Trace(r.RemoteAddr)
				log.Trace(r.Header.Get("Origin"))
				return true
			},
		},
	}

	s.initRoutes()
	http.Handle(pattern, s.corsMiddleware(s.router))

	return s
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) RunTLS(certFile string, keyFile string) error {
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

func (s *Server) WithPort(port int) *Server {
	s.server.Addr = fmt.Sprintf(":%d", port)
	return s
}

func (s *Server) WithCache(cache *cache.Cache) *Server {
	s.cache = cache
	return s
}

func (s *Server) WithTLS(config *tls.Config) *Server {
	s.server.TLSConfig = config
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
