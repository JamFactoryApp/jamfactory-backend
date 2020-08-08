package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/types"
	"jamfactory-backend/utils"
	"net/http"
	"os"
)

const (
	afterLogoutRedirect   = apiPath + authPath + authCurrentPath
	afterCallbackRedirect = "/"
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

	session.Values[models.SessionTokenKey] = token
	session.Values[models.SessionUserTypeKey] = models.UserTypeNew

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

	http.Redirect(w, r, url, http.StatusSeeOther)
	//res := loginResponseBody{Url: url}
	//utils.EncodeJSONBody(w, res)
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.logout")

	session := utils.SessionFromRequestContext(r)

	session.Options.MaxAge = -1
	SaveSession(w, r, session)
	http.Redirect(w, r, afterLogoutRedirect, http.StatusSeeOther)
}

func current(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.current")

	session := utils.SessionFromRequestContext(r)

	res := types.StatusResponseBody{}

	if session.Values[models.SessionUserTypeKey] == nil {
		res.User = models.UserTypeNew
	} else {
		res.User = session.Values[models.SessionUserTypeKey].(string)
	}

	if session.Values[models.SessionLabelTypeKey] == nil {
		res.Label = ""
	} else {
		res.Label = session.Values[models.SessionLabelTypeKey].(string)
	}

	if session.Values[models.SessionTokenKey] == nil {
		res.Authorized = false
	} else {
		token, err := utils.ParseTokenFromSession(session)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.WithField("Session", session.ID).Error("Couldn't get token: ", err.Error())
			return
		}
		res.Authorized = token.Valid()
	}

	utils.EncodeJSONBody(w, res)
}
