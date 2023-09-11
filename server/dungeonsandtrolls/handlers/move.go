package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func validateMove(game *dungeonsandtrolls.Game, c *api.Coordinates, p *gameobject.Player) error {
	if c.Level != nil && p.Position.Level != c.Level {
		return fmt.Errorf("it is not possible to travel to another level directly, please use the stairs")
	}
	// TODO check if visible
	// TODO check is free (using pathfinding)

	return nil
}

func Move(game *dungeonsandtrolls.Game, c *api.Coordinates, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	err = validateMove(game, c, p)
	if err != nil {
		return err
	}

	// TODO check that tile is free and visible
	pc := game.GetPlayerCommands(p.Character.Id)
	pc.Move = c

	return nil
}
