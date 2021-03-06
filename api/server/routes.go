package server

import (
	"github.com/gorilla/mux"
	"net/http"
)

const (
	apiPath = "/api/v1"

	authPath         = "/auth"
	authCallbackPath = "/callback"
	authLoginPath    = "/login"
	authLogoutPath   = "/logout"
	authCurrentPath  = "/current"

	jamSessionPath         = "/jam"
	jamSessionIndexPath    = ""
	jamSessionCreatePath   = "/create"
	jamSessionJoinPath     = "/join"
	jamSessionLeavePath    = "/leave"
	jamSessionPlaybackPath = "/playback"

	queuePath           = "/queue"
	queueIndexPath      = ""
	queueCollectionPath = "/collection"
	queueVotePath       = "/vote"
	queueDeletePath     = "/delete"

	spotifyPath         = "/spotify"
	spotifyDevicesPath  = "/devices"
	spotifyPlaylistPath = "/playlists"
	spotifySearchPath   = "/search"

	websocketPath = "/ws"
)

func (s *Server) initRoutes() {
	s.router.Use(s.sessionRequired)

	authRouter := s.router.PathPrefix(apiPath + authPath).Subrouter()
	jamSessionRouter := s.router.PathPrefix(apiPath + jamSessionPath).Subrouter()
	queueRouter := s.router.PathPrefix(apiPath + queuePath).Subrouter()
	spotifyRouter := s.router.PathPrefix(apiPath + spotifyPath).Subrouter()
	websocketRouter := s.router.PathPrefix(websocketPath).Subrouter()

	s.registerAuthRoutes(authRouter)
	s.registerQueueRoutes(queueRouter)
	s.registerJamSessionRoutes(jamSessionRouter)
	s.registerSpotifyRoutes(spotifyRouter)
	s.registerWebsocketRoutes(websocketRouter)
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.HandleFunc(authCallbackPath, s.callback).Methods("GET")
	r.HandleFunc(authLoginPath, s.login).Methods("GET")
	r.HandleFunc(authLogoutPath, s.logout).Methods("GET")
	r.HandleFunc(authCurrentPath, s.current).Methods("GET")
}

func (s *Server) registerJamSessionRoutes(r *mux.Router) {
	r.HandleFunc(jamSessionCreatePath, s.createJamSession).Methods("GET")
	r.HandleFunc(jamSessionJoinPath, s.joinJamSession).Methods("PUT")
	r.Handle(jamSessionLeavePath, s.jamSessionRequired(http.HandlerFunc(s.leaveJamSession))).Methods("GET")
	r.Handle(jamSessionIndexPath, s.jamSessionRequired(http.HandlerFunc(s.getJamSession))).Methods("GET")
	r.Handle(jamSessionIndexPath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.setJamSession)))).Methods("PUT")
	r.Handle(jamSessionPlaybackPath, s.jamSessionRequired(http.HandlerFunc(s.getPlayback))).Methods("GET")
	r.Handle(jamSessionPlaybackPath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.setPlayback)))).Methods("PUT")
}

func (s *Server) registerQueueRoutes(r *mux.Router) {
	r.Handle(queueIndexPath, s.jamSessionRequired(http.HandlerFunc(s.getQueue))).Methods("GET")
	r.Handle(queueCollectionPath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.addCollection)))).Methods("PUT")
	r.Handle(queueVotePath, s.jamSessionRequired(http.HandlerFunc(s.vote))).Methods("PUT")
	r.Handle(queueDeletePath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.deleteSong)))).Methods("DELETE")
}

func (s *Server) registerSpotifyRoutes(r *mux.Router) {
	r.Handle(spotifyDevicesPath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.devices)))).Methods("GET")
	r.Handle(spotifyPlaylistPath, s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.playlist)))).Methods("GET")
	r.Handle(spotifySearchPath, s.jamSessionRequired(http.HandlerFunc(s.search))).Methods("PUT")
}

func (s *Server) registerWebsocketRoutes(r *mux.Router) {
	r.Handle("", s.jamSessionRequired(http.HandlerFunc(s.websocketHandler))).Methods("GET")
}
