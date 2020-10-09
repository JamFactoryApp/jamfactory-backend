package controllers

import (
	"context"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/zmb3/spotify"
	"math/rand"
)

var jamSessions map[string]*models.JamSession

func initFactory() {
	jamSessions = make(map[string]*models.JamSession)
}

func GenerateNewJamSession(client spotify.Client) (*models.JamSession, error) {
	label := GenerateJamLabel()
	jamSession, err := models.NewJamSession(client, label)
	if err != nil {
		return nil, err
	}

	jamSessions[jamSession.Label] = jamSession

	go jamSession.Conductor()
	return jamSession, nil
}

func GenerateJamLabel() string {
	labelSlice := make([]byte, 5)
	for i := 0; i < 5; i++ {
		labelSlice[i] = models.JamLabelChars[rand.Intn(len(models.JamLabelChars))]
	}
	label := string(labelSlice)

	for _, jamSession := range jamSessions {
		if jamSession.Label == label {
			return GenerateJamLabel()
		}
	}

	return label
}

func GetJamSession(label string) *models.JamSession {
	if jamSession, exists := jamSessions[label]; exists {
		return jamSession
	}
	return nil
}

func DeleteJamSession(label string) {
	if jamSession, exists := jamSessions[label]; exists {
		ctx, cancel := context.WithCancel(jamSession.Context)
		defer cancel()

		jamSession.Context = ctx
		jamSession.SetJamSessionState(false)
		delete(jamSessions, label)
	}
}
