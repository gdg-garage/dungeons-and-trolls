package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func Commands(game *dungeonsandtrolls.Game, c *api.CommandsBatch, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	if p.Stun().IsStunned {
		return fmt.Errorf("you are stunned")
	}

	if p.IsAdmin {
		return fmt.Errorf("admin players are are not allowed to call non-monster commands")
	}

	if c.Buy != nil {
		err = dungeonsandtrolls.ValidateBuy(game, p, c.Buy)
		if err != nil {
			return err
		}
	}
	if c.Yell != nil {
		err = validateYell(game, c.Yell, p)
		if err != nil {
			return err
		}
	}
	if c.PickUp != nil {
		err = validatePickUp(game, c.PickUp, p)
		if err != nil {
			return err
		}
	}
	if c.Skill != nil {
		err = validateSkill(game, c.Skill, p)
		if err != nil {
			return err
		}
	}
	if c.AssignSkillPoints != nil {
		err = validateAssignAttributes(p, c.AssignSkillPoints)
		if err != nil {
			return err
		}
	}
	// TODO player lock
	if c.Move != nil {
		err = validateAndSetMove(game, c.Move, p)
		if err != nil {
			return err
		}
	}

	pc := game.GetCommands(p.Character.Id)
	game.CommandsLock.Lock()
	pc.Yell = c.Yell
	pc.Buy = c.Buy
	pc.PickUp = c.PickUp
	pc.Skill = c.Skill
	pc.AssignSkillPoints = c.AssignSkillPoints
	game.CommandsLock.Unlock()

	return nil
}
