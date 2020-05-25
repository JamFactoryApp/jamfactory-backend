package controllers

import (
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"jamfactory-backend/models"
)

const (
	apiPath = "/api"

	authPath         = "/auth"
	authCallbackPath = "/callback"
	authLoginPath    = "/login/"
	authLogoutPath   = "/logout"
	authStatusPath   = "/status/"

	partyPath         = "/party"
	partyIndexPath    = "/"
	partyCreatePath   = "/create"
	partyJoinPath     = "/join"
	partyLeavePath    = "/leave"
	partyPlaybackPath = "/playback"

	queuePath         = "/queue"
	queueIndexPath    = "/"
	queuePlaylistPath = "/playlist"
	queueVotePath     = "/vote"

	spotifyPath         = "/spotify"
	spotifyDevicesPath  = "/devices"
	spotifyPlaylistPath = "/playlist"
	spotifySearchPath   = "/search"

	socketIOPath    = "/socket.io/"
	socketIndexPath = "/"
)

var (
	Router *mux.Router
	Socket *socketio.Server

	authRouter    *mux.Router
	partyRouter   *mux.Router
	queueRouter   *mux.Router
	spotifyRouter *mux.Router

	socketRouter *mux.Router

	sessionRequiredMiddleware = SessionRequiredMiddleware{}
	partyRequiredMiddleware   = PartyRequiredMiddleware{}
	hostRequiredMiddleware    = UserTypeRequiredMiddleware{UserType: models.UserTypeHost}

	sessionRequired alice.Chain
	partyRequired   alice.Chain
	hostRequired    alice.Chain
)

func initRoutes() {
	Router = mux.NewRouter()

	authRouter = Router.PathPrefix(apiPath + authPath).Subrouter()
	partyRouter = Router.PathPrefix(apiPath + partyPath).Subrouter()
	queueRouter = Router.PathPrefix(apiPath + queuePath).Subrouter()
	spotifyRouter = Router.PathPrefix(apiPath + spotifyPath).Subrouter()

	socketRouter = Router.PathPrefix(socketIOPath).Subrouter()

	registerAuthRoutes()
	registerQueueRoutes()
	registerPartyRoutes()
	registerSpotifyRoutes()

	registerSocketIORoutes()
}

func initMiddleWares() {
	sessionRequired = alice.New(sessionRequiredMiddleware.Handler)
	partyRequired = sessionRequired.Append(partyRequiredMiddleware.Handler)
	hostRequired = partyRequired.Append(hostRequiredMiddleware.Handler)
}

func registerAuthRoutes() {
	authRouter.Handle(authCallbackPath, sessionRequired.ThenFunc(callback)).Methods("GET")
	authRouter.Handle(authLoginPath, sessionRequired.ThenFunc(login)).Methods("GET")
	authRouter.Handle(authLogoutPath, sessionRequired.ThenFunc(logout)).Methods("GET")
	authRouter.Handle(authStatusPath, sessionRequired.ThenFunc(status)).Methods("GET")
}

func registerPartyRoutes() {
	partyRouter.Handle(partyCreatePath, sessionRequired.ThenFunc(createParty)).Methods("GET")
	partyRouter.Handle(partyJoinPath, partyRequired.ThenFunc(joinParty)).Methods("PUT")
	partyRouter.Handle(partyLeavePath, partyRequired.ThenFunc(leaveParty)).Methods("GET")
	partyRouter.Handle(partyIndexPath, partyRequired.ThenFunc(getParty)).Methods("GET")
	partyRouter.Handle(partyIndexPath, hostRequired.ThenFunc(setParty)).Methods("PUT")
	partyRouter.Handle(partyPlaybackPath, partyRequired.ThenFunc(getPlayback)).Methods("GET")
	partyRouter.Handle(partyPlaybackPath, hostRequired.ThenFunc(setPlayback)).Methods("PUT")
}

func registerQueueRoutes() {
	queueRouter.Handle(queueIndexPath, partyRequired.ThenFunc(getQueue)).Methods("GET")
	queueRouter.Handle(queuePlaylistPath, partyRequired.ThenFunc(addPlaylist)).Methods("PUT")
	queueRouter.Handle(queueVotePath, partyRequired.ThenFunc(vote)).Methods("PUT")
}

func registerSpotifyRoutes() {
	spotifyRouter.Handle(spotifyDevicesPath, partyRequired.ThenFunc(devices)).Methods("GET")
	spotifyRouter.Handle(spotifyPlaylistPath, partyRequired.ThenFunc(playlist)).Methods("GET")
	spotifyRouter.Handle(spotifySearchPath, partyRequired.ThenFunc(search)).Methods("PUT")
}

func registerSocketIORoutes() {
	Socket.OnConnect(socketIndexPath, socketIOConnect)
	Socket.OnError(socketIndexPath, socketIOError)
	Socket.OnDisconnect(socketIndexPath, socketIODisconnect)

	socketRouter.Handle(socketIndexPath, Socket)
}
