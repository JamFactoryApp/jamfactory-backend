package controller

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterAuthRoutes(router *mux.Router) {
	router.HandleFunc("/callback", callback)
	router.HandleFunc("/login", login)
	router.HandleFunc("/refresh", refresh)
	router.HandleFunc("/status", status)
}

func callback(w http.ResponseWriter, r *http.Request) {

}

func login(w http.ResponseWriter, r *http.Request) {

}

func refresh(w http.ResponseWriter, r *http.Request) {

}

func status(w http.ResponseWriter, r *http.Request) {

}
