package models

type Env struct {
	DB DB
	Store *Sessionstore
	PartyController *PartyController
}
