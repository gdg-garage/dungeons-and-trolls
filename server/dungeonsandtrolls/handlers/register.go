package handlers

import (
	"errors"
	"fmt"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/discord"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
)

func validateRegistration(game *dungeonsandtrolls.Game, username string) error {
	if len(username) == 0 {
		return errors.New("username not provided")
	}

	if _, ok := game.Players[username]; ok {
		return errors.New("username is already used")
	}
	return nil
}

func generateApiKey() string {
	return gameobject.GetNewId()
}

func RegisterUser(game *dungeonsandtrolls.Game, user *api.User) (*api.Registration, error) {
	// TODO
	//game.GameLock.Lock()
	//defer game.GameLock.Unlock()
	userHandle, _, _ := discord.ParseUsernameAndDiscriminatorFromHandle(user.Username)
	err := validateRegistration(game, userHandle)
	if err != nil {
		return nil, err
	}
	apiKey := generateApiKey()
	err = discord.SendAPIKeyToUser(apiKey, user.Username)
	if err != nil {
		log.Warn().Err(err).Msgf("failed to send api key to %s, Discord used probably does not exist", user.Username)
		return nil, fmt.Errorf("failed to send api key to %s, Discord used probably does not exist", userHandle)
	}
	r := &api.Registration{
		ApiKey: &apiKey,
	}
	game.AddPlayer(gameobject.CreatePlayer(user.Username), r)
	r.ApiKey = nil
	return r, nil
}
