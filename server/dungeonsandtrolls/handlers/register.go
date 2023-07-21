package handlers

import (
	"errors"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func validateRegistration(game *dungeonsandtrolls.Game, user *api.User) error {
	if len(user.Username) == 0 {
		return errors.New("username not provided")
	}

	if _, ok := game.Players[user.Username]; ok {
		return errors.New("username is already used")
	}
	return nil
}

func generateApiKey() string {
	return gameobject.GetNewId()
}

func RegisterUser(game *dungeonsandtrolls.Game, user *api.User) (*api.Registration, error) {
	// TODO lock game for this (to prevent raise)
	err := validateRegistration(game, user)
	if err != nil {
		return nil, err
	}
	apiKey := generateApiKey()
	r := &api.Registration{
		ApiKey: &apiKey,
	}
	game.AddPlayer(gameobject.CreatePlayer(user.Username), r)
	return r, nil
}
