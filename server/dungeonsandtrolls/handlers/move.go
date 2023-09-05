package handlers

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func Move(game *dungeonsandtrolls.Game, c *api.Coordinates) error {
	p, err := game.GetCurrentPlayer()
	if err != nil {
		return err
	}

	// TODO check that tile is free and visible
	pc := game.GetPlayerCommands(p.Character.Id)
	pc.Move = c

	return nil
}
