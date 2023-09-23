package handlers

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func Commands(game *dungeonsandtrolls.Game, c *api.CommandsBatch, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	if c.Move != nil {
		err = validateMove(game, c.Move, p)
		if err != nil {
			return err
		}
	}
	if c.Buy != nil {
		err = dungeonsandtrolls.ValidateBuy(game, p, c.Buy)
		if err != nil {
			return err
		}
	}
	if c.Yell != nil {
		err = validateYell(game, c.Yell)
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

	pc := game.GetCommands(p.Character.Id)
	pc.Move = c.Move
	pc.Yell = c.Yell
	pc.Buy = c.Buy
	pc.PickUp = c.PickUp
	pc.Skill = c.Skill
	pc.AssignSkillPoints = c.AssignSkillPoints

	return nil
}
