package utils

import (
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"jamfactory-backend/models"
)

func ParseTokenFromSession(session *sessions.Session) *oauth2.Token {
	tokenMap := session.Values[models.SessionTokenKey].(map[string]interface{})
	token := oauth2.Token{
		AccessToken:  tokenMap["accesstoken"].(string),
		TokenType:    tokenMap["tokentype"].(string),
		RefreshToken: tokenMap["refreshtoken"].(string),
		Expiry:       tokenMap["expiry"].(primitive.DateTime).Time(),
	}
	return &token
}
