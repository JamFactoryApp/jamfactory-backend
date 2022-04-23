package server

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

const (
	api = "/api/v1"

	auth         = "/auth"
	authCallback = "/callback"
	authLogin    = "/login"
	authLogout   = "/logout"

	user          = "/me"
	userIndex     = ""
	userPlayback  = "/playback"
	userDevices   = "/devices"
	userPlaylists = "/playlists"

	jamSession         = "/jam"
	jamSessionIndex    = ""
	jamSessionCreate   = "/create"
	jamSessionJoin     = "/join"
	jamSessionLeave    = "/leave"
	jamSessionPlay     = "/play"
	jamSessionPlayback = "/playback"
	jamSessionMembers  = "/members"
	jamSessionSearch   = "/search"

	queue           = "/queue"
	queueIndex      = ""
	queueCollection = "/collection"
	queueVote       = "/vote"
	queueDelete     = "/delete"
	queueHistory    = "/history"
	queueExport     = "/export"

	spotifyIndex    = "/spotify"
	spotifyDevices  = "/devices"
	spotifyPlaylist = "/playlists"
	spotifySearch   = "/search"

	websocketPath  = "/ws"
	websocketIndex = ""
)

func (s *Server) initRoutes() {

	chain := alice.New(s.sessionMiddleware, s.userMiddleware)

	authRouter := s.router.PathPrefix(api + auth).Subrouter()
	userRouter := s.router.PathPrefix(api + user).Subrouter()
	jamSessionRouter := s.router.PathPrefix(api + jamSession).Subrouter()
	queueRouter := s.router.PathPrefix(api + queue).Subrouter()
	spotifyRouter := s.router.PathPrefix(api + spotifyIndex).Subrouter()
	websocketRouter := s.router.PathPrefix(websocketPath).Subrouter()

	s.registerAuthRoutes(authRouter, chain)
	s.registerUserRoutes(userRouter, chain)
	s.registerQueueRoutes(queueRouter, chain)
	s.registerJamSessionRoutes(jamSessionRouter, chain)
	s.registerSpotifyRoutes(spotifyRouter, chain)
	s.registerWebsocketRoutes(websocketRouter, chain)
}

func (s *Server) registerAuthRoutes(r *mux.Router, chain alice.Chain) {
	// GET: /api/v1/auth/callback
	r.Methods("GET").Path(authCallback).Handler(
		chain.Append().ThenFunc(s.callback))

	// GET: /api/v1/auth/callback
	r.Methods("GET").Path(authLogin).Handler(
		chain.Append().ThenFunc(s.login))

	// GET: /api/v1/auth/callback
	r.Methods("GET").Path(authLogout).Handler(
		chain.Append().ThenFunc(s.logout))
}

func (s *Server) registerUserRoutes(r *mux.Router, chain alice.Chain) {
	// GET: /api/v1/me
	r.Methods("GET").Path(userIndex).Handler(
		chain.Append().ThenFunc(s.getUser))

	// PUT: /api/v1/me
	r.Methods("PUT").Path(userIndex).Handler(
		chain.Append().ThenFunc(s.setUser))

	// DELETE: /api/v1/me
	r.Methods("DELETE").Path(userIndex).Handler(
		chain.Append().ThenFunc(s.deleteUser))

	// GET: /api/v1/me/playback
	r.Methods("GET").Path(userPlayback).Handler(
		chain.Append().ThenFunc(s.getUserPlayback))

	// PUT: /api/v1/me/playback
	r.Methods("PUT").Path(userPlayback).Handler(
		chain.Append().ThenFunc(s.setUserPlayback))

	// GET: /api/v1/me/devices
	r.Methods("GET").Path(userDevices).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.getUserDevices))

	// GET: /api/v1/me/playlists
	r.Methods("GET").Path(userPlaylists).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.getUserPlaylists))
}

func (s *Server) registerJamSessionRoutes(r *mux.Router, chain alice.Chain) {
	// GET: /api/v1/jam/create
	r.Methods("GET").Path(jamSessionCreate).Handler(
		chain.Append(s.nonMemberRequired).ThenFunc(s.createJamSession))

	// PUT: /api/v1/jam/join
	r.Methods("PUT").Path(jamSessionJoin).Handler(
		chain.Append(s.nonMemberRequired).ThenFunc(s.joinJamSession))

	// GET: /api/v1/jam/leave
	r.Methods("GET").Path(jamSessionLeave).Handler(
		chain.Append().ThenFunc(s.leaveJamSession))

	// PUT: /api/v1/jam/play
	r.Methods("PUT").Path(jamSessionPlay).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.playSong))

	// GET: /api/v1/jam/
	r.Methods("GET").Path(jamSessionIndex).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.getJamSession))

	// PUT: /api/v1/jam/
	r.Methods("PUT").Path(jamSessionIndex).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.setJamSession))

	// PUT: /api/v1/jam/search
	r.Methods("PUT").Path(jamSessionSearch).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.search))

	// GET: /api/v1/jam/playback
	r.Methods("GET").Path(jamSessionPlayback).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.getPlayback))

	// PUT: /api/v1/jam/playback
	r.Methods("PUT").Path(jamSessionPlayback).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.setPlayback))

	// GET: /api/v1/jam/members
	r.Methods("GET").Path(jamSessionMembers).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.getMembers))

	// PUT: /api/v1/jam/members
	r.Methods("PUT").Path(jamSessionMembers).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.setMembers))
}

func (s *Server) registerQueueRoutes(r *mux.Router, chain alice.Chain) {
	// GET: /api/v1/queue/
	r.Methods("GET").Path(queueIndex).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.getQueue))

	// PUT: /api/v1/queue/collection
	r.Methods("PUT").Path(queueCollection).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.addCollection))

	// PUT: /api/v1/queue/vote
	r.Methods("PUT").Path(queueVote).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.vote))

	// DELETE: /api/v1/queue/delete
	r.Methods("DELETE").Path(queueDelete).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.deleteSong))

	// GET: /api/v1/queue/history
	r.Methods("GET").Path(queueHistory).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.getQueueHistory))

	// PUT: /api/v1/queue/export
	r.Methods("PUT").Path(queueExport).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.exportQueue))
}

func (s *Server) registerSpotifyRoutes(r *mux.Router, chain alice.Chain) {
	// TODO: Deprecate endpoint in favour of /api/v1/user/devices
	// GET: /api/v1/spotify/devices
	r.Methods("GET").Path(spotifyDevices).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.getUserDevices))

	// TODO: Deprecate endpoint in favour of /api/v1/user/playlists
	// GET: /api/v1/spotify/playlists
	r.Methods("GET").Path(spotifyPlaylist).Handler(
		chain.Append(s.jamSessionRequired, s.hostRequired).ThenFunc(s.getUserPlaylists))

	// TODO: Deprecate endpoint in favour of /api/v1/jam/search
	// PUT: /api/v1/spotify/search
	r.Methods("PUT").Path(spotifySearch).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.search))
}

func (s *Server) registerWebsocketRoutes(r *mux.Router, chain alice.Chain) {
	// GET /ws
	r.Methods("GET").Path(websocketIndex).Handler(
		chain.Append(s.jamSessionRequired).ThenFunc(s.websocketHandler))
}
