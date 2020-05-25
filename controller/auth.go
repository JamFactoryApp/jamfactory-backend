package controller

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
	"os"
)

const (
	afterLogoutRedirect   = apiPath + authPath + authStatusPath
	afterCallbackRedirect = apiPath + authPath + authStatusPath
)

var (
	spotifyAuthenticator spotify.Authenticator
	spotifyScopes        = []string{
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
	}
)

type statusResponseBody struct {
	User  string `json:"user"`
	Label string `json:"label"`
}

type loginResponseBody struct {
	Url string `json:"url"`
}

func initSpotifyAuthenticator() {
	spotifyAuthenticator = spotify.NewAuthenticator(os.Getenv("SPOTIFY_REDIRECT_URL"), spotifyScopes...)
	spotifyAuthenticator.SetAuthInfo(os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))
}

func callback(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.callback")

	session := utils.SessionFromRequestContext(r)

	token, err := spotifyAuthenticator.Token(session.ID, r)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.WithField("Session", session.ID).Error("Couldn't get token: ", err.Error())
		return
	}

	if st := r.FormValue("state"); st != session.ID {
		http.NotFound(w, r)
		log.WithFields(log.Fields{
			"Session": session.ID,
			"State":   st,
		}).Error("State mismatch")
		return
	}

	session.Values[SessionTokenKey] = token
	session.Values[SessionUserTypeKey] = models.UserTypeHost

	SaveSession(w, r, session)
	http.Redirect(w, r, afterCallbackRedirect, http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.login")

	session := utils.SessionFromRequestContext(r)

	if session.IsNew {
		SaveSession(w, r, session)
	}

	state := session.ID
	url := spotifyAuthenticator.AuthURL(state)

	res := loginResponseBody{Url: url}
	utils.EncodeJSONBody(w, res)
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.logout")

	session := utils.SessionFromRequestContext(r)

	session.Options.MaxAge = -1
	SaveSession(w, r, session)
	http.Redirect(w, r, afterLogoutRedirect, http.StatusSeeOther)
}

func status(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.status")

	session := utils.SessionFromRequestContext(r)

	res := statusResponseBody{}

	if session.Values[SessionUserTypeKey] == nil {
		res.User = models.UserTypeNew
	} else {
		res.User = session.Values[SessionUserTypeKey].(string)
	}

	if session.Values[SessionLabelKey] == nil {
		res.Label = ""
	} else {
		res.Label = session.Values[SessionLabelKey].(string)
	}

	utils.EncodeJSONBody(w, res)
}
