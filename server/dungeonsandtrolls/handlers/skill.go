package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"github.com/rs/zerolog/log"
)

func checkDistance(playerPosition *api.Coordinates, playerAttributes *api.Attributes, c *api.Coordinates, s *api.Skill) error {
	distance := utils.ManhattanDistance(playerPosition.PositionX, playerPosition.PositionY, c.PositionX, c.PositionY)
	distanceValue, err := gameobject.AttributesValue(playerAttributes, s.Range)
	if err != nil {
		return err
	}
	distanceValue = gameobject.RoundRange(distanceValue)
	if float64(distance) > distanceValue {
		return fmt.Errorf("cast location is too far away %d > %f", distance, distanceValue)
	}
	return nil
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

	if s.Flags != nil {
		if s.Flags.Passive {
			return fmt.Errorf("passive skills cannot be used (they are used automatically)")
		}

		if s.Flags.RequiresLineOfSight {
			var targetPos gameobject.PlainPos
			switch s.Target {
			case api.Skill_character:
				ti, err := game.GetObjectById(*skillUse.TargetId)
				if err != nil {
					return fmt.Errorf("targetId %s is not valid", *skillUse.TargetId)
				}
				t, ok := ti.(gameobject.Skiller)
				if !ok {
					return fmt.Errorf("using skill on wrong object type with id %s", *skillUse.TargetId)
				}
				targetPos = gameobject.PlainPosFromApiPos(gameobject.CoordinatesToPosition(t.GetPosition()))
			case api.Skill_position:
				targetPos = gameobject.PlainPosFromApiPos(skillUse.Position)
			case api.Skill_none:
				targetPos = gameobject.PlainPosFromApiPos(gameobject.CoordinatesToPosition(p.GetPosition()))
			}

			var currentLevel *api.Level
			for _, l := range game.Game.Map.Levels {
				if l.Level == p.GetPosition().Level {
					currentLevel = l
				}
			}

			log.Info().Msgf("current level for los %+v", currentLevel)

			resultMap := make(map[gameobject.PlainPos]gameobject.MapCellExt)
			for _, objects := range currentLevel.Objects {
				resultMap[gameobject.PlainPosFromApiPos(objects.Position)] = gameobject.MapCellExt{
					MapObjects:  objects,
					Distance:    -1,
					LineOfSight: false,
				}
				log.Info().Msgf("map cell los %+v", resultMap[gameobject.PlainPosFromApiPos(objects.Position)])
			}

			//if !gameobject.GetLoS(currentLevel, resultMap, map[float32]float32{}, gameobject.CoordinatesToPosition(p.GetPosition()), targetPos) {
			//	return fmt.Errorf("target is not in los")
			//}
		}
		if s.Flags.RequiresOutOfCombat {
			if p.GetLastDamageTaken() < 3 {
				return fmt.Errorf("cannot use this skill, you have taken damage recently (out of combat flag)")
			}
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

	if skillUse.TargetId != nil {
		t, err := game.GetObjectById(*skillUse.TargetId)
		if err != nil {
			return fmt.Errorf("targetId %s is not valid", *skillUse.TargetId)
		}
		switch v := t.(type) {
		case gameobject.Skiller:
			if s.Target != api.Skill_character {
				return fmt.Errorf("the skill %s is not supposed to be used on characters", skillUse.SkillId)
			}
			err = checkDistance(p.GetPosition(), p.GetAttributes(), v.GetPosition(), s)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("using skill on wrong object type with id %s", *skillUse.TargetId)
		}
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
		if s.CasterEffects.Flags.Movement {
			pos, err := game.GetObjectsOnPosition(gameobject.PositionToCoordinates(skillUse.Position, p.GetPosition().Level))
			if err != nil {
				return fmt.Errorf("an error occured during movement %s", err.Error())
			}
			if pos != nil && !pos.IsFree {
				return fmt.Errorf("move postion (%d, %d) is not free", skillUse.Position.PositionX, skillUse.Position.PositionY)
			}
		}
		if err != nil {
			return err
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
