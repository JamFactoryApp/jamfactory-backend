package users

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/internal/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/authenticator"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"os"
)

var (
	ErrDeviceNotActive = errors.New("device not active")
)

type player struct {
	CurrentSong   *spotify.FullTrack
	Synchronized  bool
	SyncCount     int
	Active        bool
	client        spotify.Client
	spotifyPlayer *spotify.PlayerState
}

func NewPlayer(authenticator *authenticator.Authenticator, token *oauth2.Token) player {
	client := authenticator.NewClient(token)
	playerState, err := client.PlayerState()
	if err != nil {
		log.Warn(err)
	}
	return player{
		CurrentSong:   nil,
		Synchronized:  false,
		SyncCount:     0,
		Active:        false,
		client:        client,
		spotifyPlayer: playerState,
	}
}

func (p *player) Client() *spotify.Client {
	return &p.client
}

func (p *player) SetState(state bool) error {

	playerState, err := p.Client().PlayerState()
	if err != nil {
		return err
	}

	if state == playerState.Playing {
		return nil
	}

	if state {
		err = p.Client().Play()
	} else {
		err = p.Client().Pause()
	}

	if err != nil {
		return err
	}

	p.Synchronized = false
	return nil
}

func (p *player) Play(track *spotify.FullTrack) error {
	if !p.spotifyPlayer.Device.Active {
		return ErrDeviceNotActive
	}

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{track.URI},
	}

	p.Synchronized = false
	err := p.Client().PlayOpt(&playOptions)
	if err != nil {
		return err
	}
	p.CurrentSong = track

	return nil
}

func (p *player) SetDevice(id string) error {
	playerState, err := p.Client().PlayerState()
	if err != nil {
		return err
	}
	deviceID := spotify.ID(id)
	if deviceID != playerState.Device.ID {
		err := p.Client().TransferPlayback(deviceID, p.Active)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *player) Devices() ([]spotify.PlayerDevice, error) {
	return p.Client().PlayerDevices()
}

func (p *player) GetPlayerState() *spotify.PlayerState {
	return p.spotifyPlayer
}

func (p *player) SetPlayerState(state *spotify.PlayerState) {
	p.spotifyPlayer = state
}

func (p *player) Playlists() (*spotify.SimplePlaylistPage, error) {
	return p.Client().CurrentUsersPlaylists()
}

func (p *player) Search(index string, searchType spotify.SearchType, options *spotify.Options) (interface{}, error) {
	return p.Client().SearchOpt(index, searchType, options)
}

func (p *player) GetTrack(trackID string) (*spotify.FullTrack, error) {
	return p.Client().GetTrack(spotify.ID(trackID))
}

func (p *player) SetVolume(percent int) error {
	err := p.Client().Volume(percent)
	if err != nil {
		return err
	}
	return nil
}

func (p *player) CreatePlaylist(name string, desc string, ids []spotify.ID) error {
	user, err := p.Client().CurrentUser()
	if err != nil {
		return err
	}
	playlist, err := p.Client().CreatePlaylistForUser(user.ID, name, desc, false)
	if err != nil {
		return err
	}
	if utils.FileExists("./assets/playlist_cover.png") {
		file, err := os.Open("./assets/playlist_cover.png")
		defer utils.CloseProperly(file)
		if err != nil {
			return err
		}
		err = p.Client().SetPlaylistImage(playlist.ID, file)
		if err != nil {
			return err
		}
	}

	idChunks := utils.SplitsIds(ids, 100)
	for i := range idChunks {
		_, err := p.Client().AddTracksToPlaylist(playlist.ID, idChunks[i]...)
		if err != nil {
			return err
		}
	}
	return nil

}
