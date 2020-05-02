package controller

import (
	"github.com/gorilla/mux"
	"jamfactory-backend/models"
	"net/http"
)

type AuthEnv struct {
	*models.Env
}

func RegisterAuthRoutes(router *mux.Router, mainEnv *models.Env) {
	env := AuthEnv{mainEnv}
	router.HandleFunc("/callback", env.callback)
	router.HandleFunc("/login", env.login)
	router.HandleFunc("/refresh", env.refresh)
	router.HandleFunc("/status", env.status)
}

func (env *AuthEnv) callback(w http.ResponseWriter, r *http.Request) {

}

func (env *AuthEnv) login(w http.ResponseWriter, r *http.Request) {

}

func (env *AuthEnv) refresh(w http.ResponseWriter, r *http.Request) {

}

func (env *AuthEnv) status(w http.ResponseWriter, r *http.Request) {

}
