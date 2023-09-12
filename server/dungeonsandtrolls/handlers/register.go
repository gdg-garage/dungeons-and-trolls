package handlers

import (
	"errors"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/discord"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
)

func validateRegistration(game *dungeonsandtrolls.Game, user *api.User) error {
	if len(user.Username) == 0 {
		return errors.New("username not provided")
	}

	if _, ok := game.Players[user.Username]; ok {
		return errors.New("username is already used")
	}
	// TODO validate the Discord user exists.
	return nil
}

func generateApiKey() string {
	return gameobject.GetNewId()
}

func RegisterUser(game *dungeonsandtrolls.Game, user *api.User) (*api.Registration, error) {
	// TODO
	//game.GameLock.Lock()
	//defer game.GameLock.Unlock()
	err := validateRegistration(game, user)
	if err != nil {
		return nil, err
	}
	apiKey := generateApiKey()
	r := &api.Registration{
		ApiKey: &apiKey,
	}
	game.AddPlayer(gameobject.CreatePlayer(user.Username), r)
	err = discord.SendAPIKeyToUser(apiKey, user.Username)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to send api key to %s, Discord used probably does not exist", user.Username)
	}
	return r, nil
}
