package controller

import (
	"github.com/gorilla/mux"
	chain "github.com/justinas/alice"
	"github.com/zmb3/spotify"
	"jamfactory-backend/helpers"
	"jamfactory-backend/middelwares"
	"jamfactory-backend/models"
	"net/http"
)

func RegisterPartyRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	getPartyMiddleware := middelwares.GetPartyFromSession{PartyControl: &Factory}
	validateHostUserMiddleware := middelwares.ValidateUserType{User: models.UserTypeHost}

	stdChain := chain.New(getSessionMiddleware.Handler, getPartyMiddleware.Handler)
	stdHostChain := stdChain.Append(validateHostUserMiddleware.Handler)

	router.Handle("/", stdChain.ThenFunc(getParty)).Methods("GET")
	router.Handle("/", stdHostChain.ThenFunc(setParty)).Methods("PUT")
	router.Handle("/playback", stdChain.ThenFunc(getPlayback)).Methods("GET")
	router.Handle("/playback", stdHostChain.ThenFunc(setPlayback)).Methods("PUT")
}

type partyBody struct {
	Name     string     `json:"name"`
	DeviceID spotify.ID `json:"device"`
	IpVoting bool       `json:"ip"`
}

type playbackBody struct {
	CurrentSong *spotify.FullTrack   `json:"currentSong"`
	Playback    *spotify.PlayerState `json:"playback"`
}

func getParty(w http.ResponseWriter, r *http.Request) {

	party := r.Context().Value("Party").(*models.Party)

	res := partyBody{
		Name:     party.User.DisplayName,
		DeviceID: party.DeviceID,
		IpVoting: party.IpVoteEnabled,
	}

	helpers.RespondWithJSON(w, res)
}

func setParty(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	var body partyBody
	if err := helpers.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party.User.DisplayName = body.Name

	settings := models.PartySettings{
		DeviceId:  body.DeviceID,
		IpVoting:  body.IpVoting,
		PartyName: body.Name,
	}

	party.SetSetting(settings)

	res := partyBody{
		Name:     party.User.DisplayName,
		DeviceID: party.DeviceID,
		IpVoting: party.IpVoteEnabled,
	}

	helpers.RespondWithJSON(w, res)
}

func getPlayback(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	res := playbackBody{
		CurrentSong: party.CurrentSong,
		Playback:    party.PlaybackState,
	}

	helpers.RespondWithJSON(w, res)
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	var body playbackBody
	if err := helpers.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party.SetPartyState(body.Playback.Playing)

	res := playbackBody{
		CurrentSong: party.CurrentSong,
		Playback:    party.PlaybackState,
	}

	helpers.RespondWithJSON(w, res)
}
