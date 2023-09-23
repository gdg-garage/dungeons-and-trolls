package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func validateAssignAttributes(p *gameobject.Player, a *api.Attributes) error {
	// TODO should this be int?

	if a.Strength != nil && *a.Strength < 0 {
		return fmt.Errorf("strength is <0")
	}
	if a.Dexterity != nil && *a.Dexterity < 0 {
		return fmt.Errorf("dexterity is <0")
	}
	if a.Intelligence != nil && *a.Intelligence < 0 {
		return fmt.Errorf("intelligence is <0")
	}
	if a.Willpower != nil && *a.Willpower < 0 {
		return fmt.Errorf("willpower is <0")
	}
	if a.Constitution != nil && *a.Constitution < 0 {
		return fmt.Errorf("constitution is <0")
	}
	if a.SlashResist != nil && *a.SlashResist < 0 {
		return fmt.Errorf("slashResist is <0")
	}
	if a.PierceResist != nil && *a.PierceResist < 0 {
		return fmt.Errorf("pierceResist is <0")
	}
	if a.FireResist != nil && *a.FireResist < 0 {
		return fmt.Errorf("fireResist is <0")
	}
	if a.PoisonResist != nil && *a.PoisonResist < 0 {
		return fmt.Errorf("poisonResist is <0")
	}
	if a.ElectricResist != nil && *a.ElectricResist < 0 {
		return fmt.Errorf("electricResist is <0")
	}
	if a.Life != nil && *a.Life < 0 {
		return fmt.Errorf("life is <0")
	}
	if a.Stamina != nil && *a.Stamina < 0 {
		return fmt.Errorf("stamina is <0")
	}
	if a.Mana != nil && *a.Mana < 0 {
		return fmt.Errorf("mana is <0")
	}
	if a.Constant != nil && *a.Constant < 0 {
		return fmt.Errorf("scalar is <0")
	}
	if a.Constant != nil && *a.Constant != 0 {
		return fmt.Errorf("scalar cannot be changed")
	}

	sum, err := gameobject.SumAttributes(a)
	if err != nil {
		return err
	}
	if sum > float32(p.Character.SkillPoints) {
		return fmt.Errorf("not enough skill points %f > %f", sum, p.Character.SkillPoints)
	}

	return nil
}

func AssignAttributes(game *dungeonsandtrolls.Game, a *api.Attributes, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	err = validateAssignAttributes(p, a)
	if err != nil {
		return err
	}

	pc := game.GetPlayerCommands(p.Character.Id)
	pc.AssignSkillPoints = a

	return nil
}
