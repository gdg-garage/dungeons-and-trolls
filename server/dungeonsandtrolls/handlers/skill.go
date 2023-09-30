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

func tilesInRange(game *dungeonsandtrolls.Game, startingPosition *api.Coordinates, rng int32) []*api.MapObjects {
	// TODO implement
	// <=
	return []*api.MapObjects{}
}

func validateSkill(game *dungeonsandtrolls.Game, skillUse *api.SkillUse, p gameobject.Skiller) error {
	p.SetMovingTo(nil)
	if p.IsStunned() {
		return fmt.Errorf("you are stunned")
	}

	s, ok := p.GetSkill(skillUse.SkillId)
	if !ok {
		return fmt.Errorf("skill %s not found for Character %s", skillUse.SkillId, p.GetId())
	}

	if skillUse.TargetId != nil && skillUse.Position != nil {
		return fmt.Errorf("cannot use skill on target and location at the same time")
	}
	if skillUse.TargetId == nil && (s.Target == api.Skill_character) {
		return fmt.Errorf("skill targetId not specified")
	}
	if (skillUse.TargetId != nil || skillUse.Position != nil) && (s.Target == api.Skill_none) {
		return fmt.Errorf("skill target should be none")
	}
	// TODO check none skills - no position and no target

	if s.Flags != nil {
		if s.Flags.RequiresLineOfSight {
			// TODO check that target is in los
		}
		if s.Flags.RequiresOutOfCombat {
			if p.GetLastDamageTaken() < 3 {
				return fmt.Errorf("cannot use this skill, you have taken damage recently (out of combat flag)")
			}
		}
	}

	if skillUse.TargetId != nil {
		t, err := game.GetObjectById(*skillUse.TargetId)
		if err != nil {
			return fmt.Errorf("targetId %s is not valid", *skillUse.TargetId)
		}
		switch v := t.(type) {
		case *gameobject.Monster:
			if s.Target != api.Skill_character {
				return fmt.Errorf("the skill %s is not supposed to be used on characters", skillUse.SkillId)
			}
			err = checkDistance(p.GetPosition(), p.GetAttributes(), v.Position, s)
			if err != nil {
				return err
			}
		case *gameobject.Player:
			if s.Target != api.Skill_character {
				return fmt.Errorf("the skill %s is not supposed to be used on characters", skillUse.SkillId)
			}
			err = checkDistance(p.GetPosition(), p.GetAttributes(), v.Position, s)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("using skill on wrong object type with id %s", *skillUse.TargetId)
		}
		// TODO check flags
	}
	if skillUse.Position != nil {
		if skillUse.Position == nil && s.Target == api.Skill_position {
			return fmt.Errorf("skill location not specified")
		}
		l, err := game.GetCachedLevel(p.GetPosition().Level)
		if err != nil {
			return fmt.Errorf("level not found")
		}
		if skillUse.Position.PositionX >= l.Width && skillUse.Position.PositionY >= l.Height {
			return fmt.Errorf("skill target position (%d, %d) not found in the level", skillUse.Position.PositionX, skillUse.Position.PositionY)
		}
		err = checkDistance(p.GetPosition(), p.GetAttributes(), gameobject.PositionToCoordinates(skillUse.Position, p.GetPosition().Level), s)
		if err != nil {
			return err
		}
	}
	if s.Cost != nil {
		satisfied, err := gameobject.SatisfyingAttributes(p.GetAttributes(), s.Cost)
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
	if p.IsAdmin {
		return fmt.Errorf("admin players are are not allowed to call non-monster commands")
	}

	err = validateSkill(game, skillUse, p)
	if err != nil {
		return err
	}

	pc := game.GetCommands(p.Character.Id)
	pc.Skill = skillUse

	return nil
}
