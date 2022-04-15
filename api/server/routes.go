package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	api = "/api/v1"

	auth         = "/auth"
	authCallback = "/callback"
	authLogin    = "/login"
	authLogout   = "/logout"

	me      = "/me"
	meIndex = ""

	jamSession         = "/jam"
	jamSessionIndex    = ""
	jamSessionCreate   = "/create"
	jamSessionJoin     = "/join"
	jamSessionLeave    = "/leave"
	jamSessionPlay     = "/play"
	jamSessionPlayback = "/playback"
	jamSessionMembers  = "/members"

	queue           = "/queue"
	queueIndex      = ""
	queueCollection = "/collection"
	queueVote       = "/vote"
	queueDelete     = "/delete"
	queueHistory    = "/history"
	queueExport     = "/export"

	spotify         = "/spotify"
	spotifyDevices  = "/devices"
	spotifyPlaylist = "/playlists"
	spotifySearch   = "/search"

	websocket      = "/ws"
	websocketIndex = ""
)

func (s *Server) initRoutes() {
	s.router.Use(s.sessionMiddleware)
	s.router.Use(s.userMiddleware)

	authRouter := s.router.PathPrefix(api + auth).Subrouter()
	meRouter := s.router.PathPrefix(api + me).Subrouter()
	jamSessionRouter := s.router.PathPrefix(api + jamSession).Subrouter()
	queueRouter := s.router.PathPrefix(api + queue).Subrouter()
	spotifyRouter := s.router.PathPrefix(api + spotify).Subrouter()
	websocketRouter := s.router.PathPrefix(websocket).Subrouter()

	s.registerAuthRoutes(authRouter)
	s.registerMeRoutes(meRouter)
	s.registerQueueRoutes(queueRouter)
	s.registerJamSessionRoutes(jamSessionRouter)
	s.registerSpotifyRoutes(spotifyRouter)
	s.registerWebsocketRoutes(websocketRouter)
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	r.Methods("GET").Path(authCallback).HandlerFunc(s.callback)
	r.Methods("GET").Path(authLogin).HandlerFunc(s.login)
	r.Methods("GET").Path(authLogout).HandlerFunc(s.logout)
}

func (s *Server) registerMeRoutes(r *mux.Router) {
	r.Methods("GET").Path(meIndex).HandlerFunc(s.getUser)
	r.Methods("PUT").Path(meIndex).HandlerFunc(s.setUser)
	r.Methods("DELETE").Path(meIndex).HandlerFunc(s.deleteUser)
}

func (s *Server) registerJamSessionRoutes(r *mux.Router) {
	r.Methods("GET").Path(jamSessionCreate).Handler(s.nonMemberRequired(http.HandlerFunc(s.createJamSession)))
	r.Methods("PUT").Path(jamSessionJoin).Handler(s.nonMemberRequired(http.HandlerFunc(s.joinJamSession)))
	r.Methods("GET").Path(jamSessionLeave).Handler(http.HandlerFunc(s.leaveJamSession))
	r.Methods("PUT").Path(jamSessionPlay).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.playSong))))
	r.Methods("GET").Path(jamSessionIndex).Handler(s.jamSessionRequired(http.HandlerFunc(s.getJamSession)))
	r.Methods("PUT").Path(jamSessionIndex).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.setJamSession))))
	r.Methods("GET").Path(jamSessionPlayback).Handler(s.jamSessionRequired(http.HandlerFunc(s.getPlayback)))
	r.Methods("PUT").Path(jamSessionPlayback).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.setPlayback))))
	r.Methods("GET").Path(jamSessionMembers).Handler(s.jamSessionRequired(http.HandlerFunc(s.getMembers)))
	r.Methods("PUT").Path(jamSessionMembers).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.setMembers))))
}

func (s *Server) registerQueueRoutes(r *mux.Router) {
	r.Methods("GET").Path(queueIndex).Handler(s.jamSessionRequired(http.HandlerFunc(s.getQueue)))
	r.Methods("PUT").Path(queueCollection).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.addCollection))))
	r.Methods("PUT").Path(queueVote).Handler(s.jamSessionRequired(http.HandlerFunc(s.vote)))
	r.Methods("DELETE").Path(queueDelete).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.deleteSong))))
	r.Methods("GET").Path(queueHistory).Handler(s.jamSessionRequired(http.HandlerFunc(s.getQueueHistory)))
	r.Methods("PUT").Path(queueExport).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.exportQueue))))
}

func (s *Server) registerSpotifyRoutes(r *mux.Router) {
	r.Methods("GET").Path(spotifyDevices).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.devices))))
	r.Methods("GET").Path(spotifyPlaylist).Handler(s.jamSessionRequired(s.hostRequired(http.HandlerFunc(s.playlist))))
	r.Methods("PUT").Path(spotifySearch).Handler(s.jamSessionRequired(http.HandlerFunc(s.search)))
}

func (s *Server) registerWebsocketRoutes(r *mux.Router) {
	r.Methods("GET").Path(websocketIndex).Handler(s.jamSessionRequired(http.HandlerFunc(s.websocketHandler)))
}
