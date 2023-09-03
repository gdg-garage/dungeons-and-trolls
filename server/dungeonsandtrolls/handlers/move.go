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

	return game.MovePlayer(p, c)
}
