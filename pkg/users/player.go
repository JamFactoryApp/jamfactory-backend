package users

import (
	"context"
	"errors"
	"github.com/jamfactoryapp/jamfactory-backend/internal/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/authenticator"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
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
	client        *spotify.Client
	spotifyPlayer *spotify.PlayerState
}

func NewPlayer(ctx context.Context, authenticator *authenticator.Authenticator, token *oauth2.Token) player {
	client := spotify.New(authenticator.Client(ctx, token))
	playerState, err := client.PlayerState(ctx)
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
	return p.client
}

func (p *player) SetState(ctx context.Context, state bool) error {
	playerState, err := p.Client().PlayerState(ctx)
	if err != nil {
		return err
	}

	if state == playerState.Playing {
		return nil
	}

	if state {
		err = p.Client().Play(ctx)
	} else {
		err = p.Client().Pause(ctx)
	}

	if err != nil {
		return err
	}

	p.Synchronized = false
	return nil
}

func (p *player) Play(ctx context.Context, track *spotify.FullTrack) error {
	if !p.spotifyPlayer.Device.Active {
		return ErrDeviceNotActive
	}

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{track.URI},
	}

	p.Synchronized = false
	err := p.Client().PlayOpt(ctx, &playOptions)
	if err != nil {
		return err
	}
	p.CurrentSong = track

	return nil
}

func (p *player) SetDevice(ctx context.Context, id string) error {
	playerState, err := p.Client().PlayerState(ctx)
	if err != nil {
		return err
	}
	deviceID := spotify.ID(id)
	if deviceID != playerState.Device.ID {
		err := p.Client().TransferPlayback(ctx, deviceID, p.Active)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *player) Devices(ctx context.Context) ([]spotify.PlayerDevice, error) {
	return p.Client().PlayerDevices(ctx)
}

func (p *player) GetPlayerState() *spotify.PlayerState {
	return p.spotifyPlayer
}

func (p *player) SetPlayerState(state *spotify.PlayerState) {
	p.spotifyPlayer = state
}

func (p *player) Playlists(ctx context.Context) (*spotify.SimplePlaylistPage, error) {
	return p.Client().CurrentUsersPlaylists(ctx)
}

func (p *player) Search(ctx context.Context, index string, searchType spotify.SearchType, options ...spotify.RequestOption) (interface{}, error) {
	return p.Client().Search(ctx, index, searchType, options...)
}

func (p *player) GetTrack(ctx context.Context, trackID string) (*spotify.FullTrack, error) {
	return p.Client().GetTrack(ctx, spotify.ID(trackID))
}

func (p *player) SetVolume(ctx context.Context, percent int) error {
	err := p.Client().Volume(ctx, percent)
	if err != nil {
		return err
	}
	return nil
}

func (p *player) CreatePlaylist(ctx context.Context, name string, desc string, ids []spotify.ID) error {
	user, err := p.Client().CurrentUser(ctx)
	if err != nil {
		return err
	}
	playlist, err := p.Client().CreatePlaylistForUser(ctx, user.ID, name, desc, false, false)
	if err != nil {
		return err
	}
	if utils.FileExists("./assets/playlist_cover.png") {
		file, err := os.Open("./assets/playlist_cover.png")
		defer utils.CloseProperly(file)
		if err != nil {
			return err
		}
		err = p.Client().SetPlaylistImage(ctx, playlist.ID, file)
		if err != nil {
			return err
		}
	}

	idChunks := utils.SplitsIds(ids, 100)
	for i := range idChunks {
		_, err := p.Client().AddTracksToPlaylist(ctx, playlist.ID, idChunks[i]...)
		if err != nil {
			return err
		}
	}
	return nil

}
