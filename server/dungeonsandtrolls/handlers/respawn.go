package handlers

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
)

func Respawn(game *dungeonsandtrolls.Game, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	game.Respawn(p)

	return nil
}
