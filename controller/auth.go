package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
	"log"
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
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := SpotifyAuthenticator.Token(session.ID, r)

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
	err = session.Save(r, w)

	if err != nil {
		log.Println("Couldn't save session")
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func login(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Println("Couldn't encode json")
	}
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Options.MaxAge = -1
	err = Store.Save(r, w, session)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func status(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		log.Println("Couldn't encode json")
	}
}
