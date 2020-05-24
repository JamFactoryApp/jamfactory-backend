package controller

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	chain "github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/helpers"
	"jamfactory-backend/middelwares"
	"jamfactory-backend/models"
	"net/http"
	"os"
)

var SpotifyAuthenticator spotify.Authenticator

func RegisterAuthRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	stdChain := chain.New(getSessionMiddleware.Handler)
	SpotifyAuthenticator.SetAuthInfo(os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))

	router.Handle("/callback", stdChain.ThenFunc(callback))
	router.Handle("/login/", stdChain.ThenFunc(login))
	router.Handle("/logout", stdChain.ThenFunc(logout))
	router.Handle("/status/", stdChain.ThenFunc(status))
}

func callback(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.callback")

	session := r.Context().Value("Session").(*sessions.Session)

	token, err := SpotifyAuthenticator.Token(session.ID, r)

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
	session.Values[models.SessionUserKey] = models.UserTypeHost

	helpers.SaveSession(w, r, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.login")

	session := r.Context().Value("Session").(*sessions.Session)

	if session.IsNew {
		helpers.SaveSession(w, r, session)
	}

	state := session.ID
	url := SpotifyAuthenticator.AuthURL(state)

	res := make(map[string]interface{})
	res["url"] = url

	helpers.RespondWithJSON(w, res)
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.logout")

	session := r.Context().Value("Session").(*sessions.Session)

	session.Options.MaxAge = -1
	helpers.SaveSession(w, r, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func status(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.status")

	session := r.Context().Value("Session").(*sessions.Session)

	res := make(map[string]interface{})

	if session.Values[models.SessionUserKey] == nil {
		res["user"] = models.UserTypeNew
	} else {
		res["user"] = session.Values[models.SessionUserKey]
	}

	if session.Values[models.SessionLabelKey] == nil {
		res["label"] = ""
	} else {
		res["label"] = session.Values[models.SessionLabelKey]
	}

	helpers.RespondWithJSON(w, res)
}
