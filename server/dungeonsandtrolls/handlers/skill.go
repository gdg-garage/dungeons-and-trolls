package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
)

func checkDistance(playerPosition *api.Coordinates, playerAttributes *api.Attributes, c *api.Coordinates, s *api.Skill) error {
	distance := utils.ManhattanDistance(playerPosition.PositionX, playerPosition.PositionY, c.PositionX, c.PositionY)
	distanceValue, err := gameobject.AttributesValue(playerAttributes, s.Range)
	if err != nil {
		return err
	}
	if float64(distance) > gameobject.RoundRange(distanceValue) {
		return fmt.Errorf("cast location is too far away")
	}
	return nil
}

func validateSkill(game *dungeonsandtrolls.Game, skillUse *api.SkillUse, p *gameobject.Player) error {
	s, ok := p.Skills[skillUse.SkillId]
	if !ok {
		return fmt.Errorf("skill %s not found for Character %s", skillUse.SkillId, p.Character.Id)
	}
	if skillUse.TargetId != nil && skillUse.Coordinates != nil {
		return fmt.Errorf("cannot use skill on target and location at the same time")
	}
	if skillUse.TargetId == nil && (s.Target == api.Skill_character || s.Target == api.Skill_item) {
		return fmt.Errorf("skill targetId not specified")
	}
	if (skillUse.TargetId != nil || skillUse.Coordinates != nil) && (s.Target == api.Skill_none) {
		return fmt.Errorf("skill target should be none")
	}
	if skillUse.TargetId != nil {
		t, err := game.GetObjectById(*skillUse.TargetId)
		if err != nil {
			return fmt.Errorf("targetId %s is not valid", *skillUse.TargetId)
		}
		switch v := t.(type) {
		case *api.Item:
			if s.Target != api.Skill_item {
				return fmt.Errorf("the skill %s is not supposed to be used on items", skillUse.SkillId)
			}
			found := false
			for _, i := range p.Equipped {
				if i.Id == *skillUse.TargetId {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("trying to cast to a non-owned item %s", *skillUse.TargetId)
			}
		case *gameobject.Monster:
			if s.Target != api.Skill_character {
				return fmt.Errorf("the skill %s is not supposed to be used on characters", skillUse.SkillId)
			}
			err = checkDistance(p.Position, p.Character.Attributes, v.Position, s)
			if err != nil {
				return err
			}
		case *gameobject.Player:
			if s.Target != api.Skill_character {
				return fmt.Errorf("the skill %s is not supposed to be used on characters", skillUse.SkillId)
			}
			err = checkDistance(p.Position, p.Character.Attributes, v.Position, s)
			if err != nil {
				return err
			}
		}
		// TODO check flags
	}
	if skillUse.Coordinates != nil {
		if skillUse.Coordinates == nil && s.Target == api.Skill_position {
			return fmt.Errorf("skill location not specified")
		}
		l, err := game.GetCachedLevel(*p.Position.Level)
		if err != nil {
			return fmt.Errorf("level not found")
		}
		if skillUse.Coordinates.PositionX >= l.Width && skillUse.Coordinates.PositionY >= l.Height {
			return fmt.Errorf("skill target position (%d, %d) not found in the level", skillUse.Coordinates.PositionX, skillUse.Coordinates.PositionY)
		}
		err = checkDistance(p.Position, p.Character.Attributes, gameobject.PositionToCoordinates(skillUse.Coordinates, *p.Position.Level), s)
		if err != nil {
			return err
		}
	}
	if s.Cost != nil {
		satisfied, err := gameobject.SatisfyingAttributes(p.Character.Attributes, s.Cost)
		if err != nil {
			return err
		}
		if !satisfied {
			return fmt.Errorf("requirements (cost) for the skill are not satisfied")
		}
	}
	return nil
}

func Skill(game *dungeonsandtrolls.Game, skillUse *api.SkillUse, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	err = validateSkill(game, skillUse, p)
	if err != nil {
		return err
	}

	pc := game.GetPlayerCommands(p.Character.Id)
	pc.Skill = skillUse

	return nil
}
