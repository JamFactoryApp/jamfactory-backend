package controllers

import (
	"github.com/gorilla/mux"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	"github.com/justinas/alice"
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

var (
	Router *mux.Router

	authRouter       *mux.Router
	jamSessionRouter *mux.Router
	queueRouter      *mux.Router
	spotifyRouter    *mux.Router

	websocketRouter *mux.Router

	sessionRequiredMiddleware    = SessionRequiredMiddleware{}
	jamSessionRequiredMiddleware = JamSessionRequiredMiddleware{}
	hostRequiredMiddleware       = UserTypeRequiredMiddleware{UserType: models.UserTypeHost}

	sessionRequired    alice.Chain
	jamSessionRequired alice.Chain
	hostRequired       alice.Chain

	cache *utils.RedisCache
)

func initRoutes() {
	Router = mux.NewRouter()

	authRouter = Router.PathPrefix(apiPath + authPath).Subrouter()
	jamSessionRouter = Router.PathPrefix(apiPath + jamSessionPath).Subrouter()
	queueRouter = Router.PathPrefix(apiPath + queuePath).Subrouter()
	spotifyRouter = Router.PathPrefix(apiPath + spotifyPath).Subrouter()

	websocketRouter = Router.PathPrefix(websocketPath).Subrouter()

	registerAuthRoutes()
	registerQueueRoutes()
	registerJamSessionRoutes()
	registerSpotifyRoutes()

	registerWebsocketRoutes()

	cache = utils.NewRedisCache(models.RedisPool.Get(), utils.RedisKey{}.Append("cache"), 30)
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
	jamSessionRouter.Handle(jamSessionJoinPath, sessionRequired.ThenFunc(joinJamSession)).Methods("PUT")
	jamSessionRouter.Handle(jamSessionLeavePath, sessionRequired.ThenFunc(leaveJamSession)).Methods("GET")
	jamSessionRouter.Handle(jamSessionIndexPath, jamSessionRequired.ThenFunc(getJamSession)).Methods("GET")
	jamSessionRouter.Handle(jamSessionIndexPath, hostRequired.ThenFunc(setJamSession)).Methods("PUT")
	jamSessionRouter.Handle(jamSessionPlaybackPath, jamSessionRequired.ThenFunc(getPlayback)).Methods("GET")
	jamSessionRouter.Handle(jamSessionPlaybackPath, hostRequired.ThenFunc(setPlayback)).Methods("PUT")
}

func registerQueueRoutes() {
	queueRouter.Handle(queueIndexPath, jamSessionRequired.ThenFunc(getQueue)).Methods("GET")
	queueRouter.Handle(queueCollectionPath, hostRequired.ThenFunc(addCollection)).Methods("PUT")
	queueRouter.Handle(queueVotePath, jamSessionRequired.ThenFunc(vote)).Methods("PUT")
	queueRouter.Handle(queueDeletePath, hostRequired.ThenFunc(deleteSong)).Methods("DELETE")
}

func registerSpotifyRoutes() {
	spotifyRouter.Handle(spotifyDevicesPath, hostRequired.ThenFunc(devices)).Methods("GET")
	spotifyRouter.Handle(spotifyPlaylistPath, hostRequired.ThenFunc(playlist)).Methods("GET")
	spotifyRouter.Handle(spotifySearchPath, jamSessionRequired.ThenFunc(search)).Methods("PUT")
}

func registerWebsocketRoutes() {
	websocketRouter.Handle("", jamSessionRequired.ThenFunc(websocketHandler)).Methods("GET")
}
