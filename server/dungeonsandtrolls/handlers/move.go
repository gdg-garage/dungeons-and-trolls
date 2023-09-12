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
	// check that path exists
	lc, err := game.GetCachedLevel(*p.Position.Level)
	if err != nil {
		return err
	}
	if c.PositionX >= lc.Width || c.PositionY >= lc.Height {
		return fmt.Errorf("position (%d, %d) is out of the level map", c.PositionX, c.PositionY)
	}
	path := lc.Grid.GetPathFromCells(
		lc.Grid.Get(int(p.Position.PositionX), int(p.Position.PositionY)),
		lc.Grid.Get(int(c.PositionX), int(c.PositionY)), false, true)
	if path == nil {
		return fmt.Errorf("there is no valid path from (%d, %d) to (%d, %d)",
			p.Position.PositionX, p.Position.PositionY, c.PositionX, c.PositionY)
	}
	if path.Length() == 0 {
		return fmt.Errorf("there is no valid path from (%d, %d) to (%d, %d)",
			p.Position.PositionX, p.Position.PositionY, c.PositionX, c.PositionY)
	}
	p.MovingTo = path
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

	return nil
}
