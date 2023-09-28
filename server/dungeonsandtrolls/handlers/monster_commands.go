package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func validateMonsterCommands(game *dungeonsandtrolls.Game, mc *api.CommandsForMonsters, p *gameobject.Player) error {
	if !p.IsAdmin {
		return fmt.Errorf("you need to be a monster puppeteer (admin) to do this")
	}

	for mId, c := range mc.Commands {
		o, err := game.GetObjectById(mId)
		if err != nil {
			return fmt.Errorf("tried to control monster with ID %s which does not exist", mId)
		}
		m, ok := o.(*gameobject.Monster)
		if m.Stun.IsStunned {
			return fmt.Errorf("tried to control stunned monster")
		}
		if !ok {
			return fmt.Errorf("tried to control %s which is not a monster", mId)
		}
		if c.Buy != nil {
			return fmt.Errorf("monsters are not allowed to shop")
		}
		if c.PickUp != nil {
			return fmt.Errorf("monsters are not allowed to pick up")
		}
		if c.AssignSkillPoints != nil {
			return fmt.Errorf("monsters are not allowed to shop")
		}
		if c.Yell != nil {
			err = validateYell(game, c.Yell)
			if err != nil {
				return err
			}
		}
		if c.Skill != nil {
			err = validateSkill(game, c.Skill, m)
			if err != nil {
				return err
			}
		}
		// TODO player lock
		if c.Move != nil {
			err = validateAndSetMove(game, c.Move, m)
			if err != nil {
				return err
			}
		}

		pc := game.GetCommands(mId)
		pc.Yell = c.Yell
		pc.Skill = c.Skill
	}
	return nil
}

func MonsterCommands(game *dungeonsandtrolls.Game, b *api.CommandsForMonsters, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	err = validateMonsterCommands(game, b, p)
	if err != nil {
		return err
	}

	return nil
}
