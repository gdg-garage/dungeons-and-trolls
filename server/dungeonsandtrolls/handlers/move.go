package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
)

func validateAndSetMove(game *dungeonsandtrolls.Game, c *api.Position, p gameobject.Alive) error {
	pc := game.GetCommands(p.GetId())
	if pc.Skill != nil {
		return fmt.Errorf("cannot move and use skill at the same time")
	}
	if p.IsStunned() {
		return fmt.Errorf("you are stunned")
	}
	// check that the destination is still the same
	if p.GetMovingTo() != nil {
		// character  is already moving there - then do nothing
		last := p.GetMovingTo().Get(p.GetMovingTo().Length() - 1)
		if last.X == int(c.PositionX) && last.Y == int(c.PositionY) {
			return nil
		}
	}
	// TODO check if visible
	// check that path exists
	lc, err := game.GetCachedLevel(p.GetPosition().Level)
	if err != nil {
		return err
	}
	if c.PositionX >= lc.Width || c.PositionY >= lc.Height {
		return fmt.Errorf("position (%d, %d) is out of the level map", c.PositionX, c.PositionY)
	}
	if p.GetPosition().PositionX >= lc.Width || p.GetPosition().PositionY >= lc.Height {
		return fmt.Errorf("player position (%d, %d) is out of the level map", p.GetPosition().PositionX, p.GetPosition().PositionY)
	}
	if p.GetPosition() == nil {
		log.Warn().Msgf("trying to move %s which is on nil pos", p.GetId())
		return fmt.Errorf("tried to move with %s which is on nil pos", p.GetId())
	}
	if lc.Grid == nil {
		log.Warn().Msgf("trying to move on level %d which does not have paths", p.GetPosition().Level)
		return fmt.Errorf("tried to move on level %d wihthout any paths", p.GetPosition().Level)
	}

	path := lc.Grid.GetPathFromCells(
		lc.Grid.Get(int(p.GetPosition().PositionX), int(p.GetPosition().PositionY)),
		lc.Grid.Get(int(c.PositionX), int(c.PositionY)), false, true)
	if path == nil {
		return fmt.Errorf("there is no valid path from (%d, %d) to (%d, %d)",
			p.GetPosition().PositionX, p.GetPosition().PositionY, c.PositionX, c.PositionY)
	}
	if path.Length() == 0 {
		return fmt.Errorf("there is no valid path from (%d, %d) to (%d, %d)",
			p.GetPosition().PositionX, p.GetPosition().PositionY, c.PositionX, c.PositionY)
	}
	p.SetMovingTo(path)
	return nil
}

func Move(game *dungeonsandtrolls.Game, c *api.Position, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}
	if p.IsAdmin {
		return fmt.Errorf("admin players are are not allowed to call non-monster commands")
	}

	err = validateAndSetMove(game, c, p)
	if err != nil {
		return err
	}

	return nil
}
