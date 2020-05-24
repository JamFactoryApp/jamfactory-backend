package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"net/http"
	"os"
)

var SpotifyAuthenticator spotify.Authenticator

func RegisterAuthRoutes(router *mux.Router) {
	SpotifyAuthenticator.SetAuthInfo(os.Getenv("SPOTIFY_ID"), os.Getenv("SPOTIFY_SECRET"))

	router.HandleFunc("/callback", callback)
	router.HandleFunc("/login/", login)
	router.HandleFunc("/logout", logout)
	router.HandleFunc("/status/", status)
}

func callback(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.callback")

	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Could not get session: ", err.Error())
		return
	}

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

	session.Values["Token"] = token
	session.Values["User"] = "Host"
	err = session.Save(r, w)

	if err != nil {
		log.WithField("Session", session.ID).Error("Couldn't save session: ", err.Error())
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.login")

	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Could not get session", err.Error())
		return
	}

	if session.IsNew {
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.WithField("Session", session.ID).Error("Could not save session", err.Error())
			return
		}
	}

	state := session.ID
	url := SpotifyAuthenticator.AuthURL(state)

	res := make(map[string]interface{})
	res["url"] = url

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		log.Error("Couldn't encode json: ", err.Error())
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.logout")

	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Could not get session: ", err.Error())
		return
	}

	session.Options.MaxAge = -1
	err = session.Save(r, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Session", session.ID).Error("Could not save session", err.Error())
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func status(w http.ResponseWriter, r *http.Request) {
	log.Trace("Controller call: auth.status")

	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Could not get session: ", err.Error())
		return
	}

	res := make(map[string]interface{})

	if session.Values["User"] == nil {
		res["user"] = "New"
		res["label"] = ""
	} else {
		if session.Values["User"] == "Guest" {
			res["user"] = "Guest"
			if session.Values["Label"] == nil {
				res["label"] = ""
			} else {
				res["label"] = session.Values["Label"].(string)
			}
		} else if session.Values["User"] == "Host" {
			res["user"] = "Host"
			if session.Values["Label"] == nil {
				res["label"] = ""
			} else {
				res["label"] = session.Values["Label"].(string)
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		log.Error("Couldn't encode json: ", err.Error())
	}
}
