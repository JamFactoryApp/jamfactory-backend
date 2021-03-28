package server

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"time"
)

func init() {
	gob.Register(&oauth2.Token{})
	gob.Register(&spotify.SearchResult{})
	gob.Register(types.UserType(""))
}

const (
	readTimeout  = time.Second
	writeTimeout = 5 * time.Second
	idleTimeout  = time.Second
)

type Server struct {
	server     *http.Server
	router     *mux.Router
	store      sessions.Store
	cache      cache.Cache
	jamFactory JamFactory
	upgrader   websocket.Upgrader
}

func NewServer(pattern string, store sessions.Store, jamFactory JamFactory) *Server {
	s := &Server{
		server: &http.Server{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		},
		router:     mux.NewRouter(),
		store:      store,
		jamFactory: jamFactory,
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

func (s *Server) WithAddress(address *url.URL) *Server {
	s.server.Addr = fmt.Sprintf(":%s", address.Port())
	return s
}

func (s *Server) WithCache(cache cache.Cache) *Server {
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
