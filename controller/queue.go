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
)

func RegisterQueueRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	getPartyMiddleware := middelwares.GetPartyFromSession{PartyControl: &Factory}
	parsePlaylistBodyMiddleware := middelwares.BodyParser{Body: new(playlistBody)}
	parseVoteBodyMiddleware := middelwares.BodyParser{Body: new(voteBody)}

	stdChain := chain.New(getSessionMiddleware.Handler, getPartyMiddleware.Handler)

	router.Handle("/", stdChain.ThenFunc(getQueue)).Methods("GET")
	router.Handle("/playlist", stdChain.Append(parsePlaylistBodyMiddleware.Handler).ThenFunc(addPlaylist)).Methods("PUT")
	router.Handle("/vote", stdChain.Append(parseVoteBodyMiddleware.Handler).ThenFunc(vote)).Methods("PUT")
}

type voteBody struct {
	Song spotify.FullTrack `json:"song"`
}

type playlistBody struct {
	PlaylistURI spotify.ID `json:"uri"`
}

func addPlaylist(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)
	body := r.Context().Value("Body").(*playlistBody)

	playlist, err := party.Client.GetPlaylistTracks(body.PlaylistURI)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Debug("Could not get playlist: ", err.Error())
		return
	}

	for i := 0; i < len(playlist.Tracks); i++ {
		party.Queue.Vote("Host", playlist.Tracks[i].Track)
	}

	queue := party.Queue.GetObjectWithoutId("")
	party.Socket.BroadcastToRoom("/", party.Label, "queue", party.Queue.GetObjectWithoutId(""))
	helpers.RespondWithJSON(w, queue)
}

func getQueue(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value("Session").(*sessions.Session)
	party := r.Context().Value("Party").(*models.Party)

	voteID := session.ID
	if party.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	queue := party.Queue.GetObjectWithoutId(voteID)
	helpers.RespondWithJSON(w, queue)
}

func vote(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value("Session").(*sessions.Session)
	party := r.Context().Value("Party").(*models.Party)
	body := r.Context().Value("Body").(*voteBody)

	voteID := session.ID
	if party.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	party.Queue.Vote(voteID, body.Song)
	queue := party.Queue.GetObjectWithoutId(voteID)
	party.Socket.BroadcastToRoom("/", party.Label, "queue", party.Queue.GetObjectWithoutId(""))
	helpers.RespondWithJSON(w, queue)
}
