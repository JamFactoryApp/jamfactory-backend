package controller

import (
	"github.com/gorilla/mux"
	"jamfactory-backend/models"
)

type QueueEnv struct {
	*models.Env
}

func RegisterQueueRoutes(router *mux.Router, mainEnv *models.Env) {
	// env := QueueEnv{mainEnv}

}
