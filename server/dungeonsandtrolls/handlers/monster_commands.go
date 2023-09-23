package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
)

func validateMonsterCommands(game *dungeonsandtrolls.Game, mc *api.CommandsForMonsters, p *gameobject.Player) error {
	if !p.IsAdmin {
		return fmt.Errorf("you need to be a monster puppeteer (admin) to do this")
	}

	for mId, c := range mc.Commands {
		o, err := game.GetObjectById(mId)
		if err != nil {
			log.Warn().Err(err).Msgf("monster admin tried to control monster with ID %s which does not exist", mId)
		}
		m, ok := o.(*gameobject.Monster)
		if !ok {
			return fmt.Errorf("tried to control %s which is not a monster", mId)
		}
		if c.Move != nil {
			err = validateMove(game, c.Move, m)
			if err != nil {
				return err
			}
		}
		if c.Buy != nil {
			return fmt.Errorf("monsters are not allowed to shop")
		}
		if c.Yell != nil {
			err = validateYell(game, c.Yell)
			if err != nil {
				return err
			}
		}
		if c.PickUp != nil {
			return fmt.Errorf("monsters are not allowed to pick up")
		}
		// TODO
		//if c.Skill != nil {
		//	err = validateSkill(game, c.Skill, m)
		//	if err != nil {
		//		return err
		//	}
		//}
		if c.AssignSkillPoints != nil {
			return fmt.Errorf("monsters are not allowed to shop")
		}
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
