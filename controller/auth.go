package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"net/http"
	"os"
)

type AuthEnv struct {
	*models.Env
	SpotifyAuthenticator *spotify.Authenticator
}

func RegisterAuthRoutes(router *mux.Router, mainEnv *models.Env) {
	spotifyAuthenticator := spotify.NewAuthenticator(os.Getenv("SPOTIFY_REDIRECT_URL"),
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState)

	env := AuthEnv{
		Env: mainEnv,
		SpotifyAuthenticator: &spotifyAuthenticator}
	router.HandleFunc("/callback", env.callback)
	router.HandleFunc("/login", env.login)
	router.HandleFunc("/status", env.status)
}

func (env *AuthEnv) callback(w http.ResponseWriter, r *http.Request) {

	session, err := env.Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	token, err := env.SpotifyAuthenticator.Token(session.ID, r)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		return
	}

	if st := r.FormValue("state"); st != session.ID {
		http.NotFound(w, r)
		return
	}

	session.Values["Token"] = token
	session.Values["User"] = "Host"
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (env *AuthEnv) login(w http.ResponseWriter, r *http.Request) {

	session, err := env.Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	state := session.ID
	url := env.SpotifyAuthenticator.AuthURL(state)

	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (env *AuthEnv) status(w http.ResponseWriter, r *http.Request) {
	session, err := env.Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	res := make(map[string]interface{})

	if session.Values["User"] != nil {
		res["user"] = "New"
		res["label"] = ""
	} else {
		if session.Values["User"] == "Guest" {
			res["user"] = "Guest"
			res["label"] = session.Values["Label"].(string)
		} else if session.Values["User"] == "Host" {
			res["user"] = "Host"
			res["label"] = session.Values["Label"].(string)
		} else {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)

}
