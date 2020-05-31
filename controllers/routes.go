package controllers

import (
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"jamfactory-backend/models"
)

const (
	apiPath = "/api/v1"

	authPath         = "/auth"
	authCallbackPath = "/callback"
	authLoginPath    = "/login"
	authLogoutPath   = "/logout"
	authCurrentPath  = "/current"

	jamSessionPath         = "/jam"
	jamSessionIndexPath    = "/"
	jamSessionCreatePath   = "/create"
	jamSessionJoinPath     = "/join"
	jamSessionLeavePath    = "/leave"
	jamSessionPlaybackPath = "/playback"

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

	authRouter       *mux.Router
	jamSessionRouter *mux.Router
	queueRouter      *mux.Router
	spotifyRouter    *mux.Router

	socketRouter *mux.Router

	sessionRequiredMiddleware    = SessionRequiredMiddleware{}
	jamSessionRequiredMiddleware = JamSessionRequiredMiddleware{}
	hostRequiredMiddleware       = UserTypeRequiredMiddleware{UserType: models.UserTypeHost}

	sessionRequired    alice.Chain
	jamSessionRequired alice.Chain
	hostRequired       alice.Chain
)

func initRoutes() {
	Router = mux.NewRouter()

	authRouter = Router.PathPrefix(apiPath + authPath).Subrouter()
	jamSessionRouter = Router.PathPrefix(apiPath + jamSessionPath).Subrouter()
	queueRouter = Router.PathPrefix(apiPath + queuePath).Subrouter()
	spotifyRouter = Router.PathPrefix(apiPath + spotifyPath).Subrouter()

	socketRouter = Router.PathPrefix(socketIOPath).Subrouter()

	registerAuthRoutes()
	registerQueueRoutes()
	registerJamSessionRoutes()
	registerSpotifyRoutes()

	registerSocketIORoutes()
}

func initMiddleWares() {
	sessionRequired = alice.New(sessionRequiredMiddleware.Handler)
	jamSessionRequired = sessionRequired.Append(jamSessionRequiredMiddleware.Handler)
	hostRequired = jamSessionRequired.Append(hostRequiredMiddleware.Handler)
}

func registerAuthRoutes() {
	authRouter.Handle(authCallbackPath, sessionRequired.ThenFunc(callback)).Methods("GET")
	authRouter.Handle(authLoginPath, sessionRequired.ThenFunc(login)).Methods("GET")
	authRouter.Handle(authLogoutPath, sessionRequired.ThenFunc(logout)).Methods("GET")
	authRouter.Handle(authCurrentPath, sessionRequired.ThenFunc(current)).Methods("GET")
}

func registerJamSessionRoutes() {
	jamSessionRouter.Handle(jamSessionCreatePath, sessionRequired.ThenFunc(createJamSession)).Methods("GET")
	jamSessionRouter.Handle(jamSessionJoinPath, jamSessionRequired.ThenFunc(joinJamSession)).Methods("PUT")
	jamSessionRouter.Handle(jamSessionLeavePath, jamSessionRequired.ThenFunc(leaveJamSession)).Methods("GET")
	jamSessionRouter.Handle(jamSessionIndexPath, jamSessionRequired.ThenFunc(getJamSession)).Methods("GET")
	jamSessionRouter.Handle(jamSessionIndexPath, hostRequired.ThenFunc(setJamSession)).Methods("PUT")
	jamSessionRouter.Handle(jamSessionPlaybackPath, jamSessionRequired.ThenFunc(getPlayback)).Methods("GET")
	jamSessionRouter.Handle(jamSessionPlaybackPath, hostRequired.ThenFunc(setPlayback)).Methods("PUT")
}

func registerQueueRoutes() {
	queueRouter.Handle(queueIndexPath, jamSessionRequired.ThenFunc(getQueue)).Methods("GET")
	queueRouter.Handle(queuePlaylistPath, jamSessionRequired.ThenFunc(addPlaylist)).Methods("PUT")
	queueRouter.Handle(queueVotePath, jamSessionRequired.ThenFunc(vote)).Methods("PUT")
}

func registerSpotifyRoutes() {
	spotifyRouter.Handle(spotifyDevicesPath, jamSessionRequired.ThenFunc(devices)).Methods("GET")
	spotifyRouter.Handle(spotifyPlaylistPath, jamSessionRequired.ThenFunc(playlist)).Methods("GET")
	spotifyRouter.Handle(spotifySearchPath, jamSessionRequired.ThenFunc(search)).Methods("PUT")
}

func registerSocketIORoutes() {
	Socket.OnConnect(socketIndexPath, socketIOConnect)
	Socket.OnError(socketIndexPath, socketIOError)
	Socket.OnDisconnect(socketIndexPath, socketIODisconnect)

	socketRouter.Handle(socketIndexPath, Socket)
}
