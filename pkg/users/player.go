package users

import (
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/internal/utils"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"os"
)

var (
	ErrDeviceNotActive = errors.New("device not active")
)

type Player struct {
	authenticator *Authenticator
	SpotifyToken  *oauth2.Token
	CurrentSong   *spotify.FullTrack
	Synchronized  bool
	SyncCount     int
	Active        bool
	client        *spotify.Client
	spotifyPlayer *spotify.PlayerState
}

func (p Player) Client() *spotify.Client {
	if p.client == nil {
		client := p.authenticator.NewClient(p.SpotifyToken)
		p.client = &client
		return p.client
	}
	return p.client
}

func (p Player) SetState(state bool) error {

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

func (p Player) Play(track *spotify.FullTrack) error {
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

func (p Player) SetDevice(id string) error {
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

func (p Player) Devices() ([]spotify.PlayerDevice, error) {
	return p.Client().PlayerDevices()
}

func (p Player) GetPlayerState() *spotify.PlayerState {
	return p.spotifyPlayer
}

func (p Player) SetPlayerState(state *spotify.PlayerState) {
	p.spotifyPlayer = state
}

func (p Player) Playlists() (*spotify.SimplePlaylistPage, error) {
	return p.Client().CurrentUsersPlaylists()
}

func (p Player) Search(index string, searchType spotify.SearchType, options *spotify.Options) (interface{}, error) {
	return p.Client().SearchOpt(index, searchType, options)
}

func (p Player) GetTrack(trackID string) (*spotify.FullTrack, error) {
	return p.Client().GetTrack(spotify.ID(trackID))
}

func (p Player) SetVolume(percent int) error {
	err := p.Client().Volume(percent)
	if err != nil {
		return err
	}
	return nil
}

func (p Player) CreatePlaylist(name string, desc string, ids []spotify.ID) error {
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
