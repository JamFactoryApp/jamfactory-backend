package utils

import (
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"jamfactory-backend/models"
)

func ParseTokenFromSession(session *sessions.Session) *oauth2.Token {
	token := session.Values[models.SessionTokenKey].(oauth2.Token)
	//token := oauth2.Token{
	//	AccessToken:  tokenMap["accesstoken"].(string),
	//	TokenType:    tokenMap["tokentype"].(string),
	//	RefreshToken: tokenMap["refreshtoken"].(string),
	//	Expiry:       tokenMap["expiry"].(primitive.DateTime).Time(),
	//}
	return &token
}
