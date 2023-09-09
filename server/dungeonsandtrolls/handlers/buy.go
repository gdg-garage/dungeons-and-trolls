package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func Buy(game *dungeonsandtrolls.Game, identifiers *api.Identifiers, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	if p.Position.Level == &gameobject.ZeroLevel {
		return fmt.Errorf("buying is availible only on the ground floor")
	}

	err = dungeonsandtrolls.ValidateBuy(game, identifiers, p)
	if err != nil {
		return err
	}

	pc := game.GetPlayerCommands(p.Character.Id)
	pc.Buy = identifiers

	return nil
}
