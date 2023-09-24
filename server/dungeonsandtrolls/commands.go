package dungeonsandtrolls

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

// ValidateBuy validates identifiers, funds, and requirements
func ValidateBuy(game *Game, p *gameobject.Player, identifiers *api.Identifiers) error {
	// this has to be a copy
	playerMoney := p.Character.Money * 1
	requirements := &api.Attributes{}
	attributes := &api.Attributes{}

	equip := make(map[api.Item_Type]*api.Item, len(p.Equipped))
	for k, v := range p.Equipped {
		equip[k] = v
	}

	for _, id := range identifiers.Ids {
		maybeItem, err := game.GetObjectById(id)
		if err != nil {
			return err
		}
		i, ok := maybeItem.(*api.Item)
		if !ok {
			return fmt.Errorf("%s is not an Item ID", id)
		}
		playerMoney -= i.Price
		if playerMoney < 0 {
			return fmt.Errorf("insufficient funds to make the purchase")
		}
		if i.Requirements != nil {
			err := gameobject.MaxAllAttributes(requirements, i.Requirements, false)
			if err != nil {
				return err
			}
		}
		equip[i.Slot] = i
	}

	// Propagate base attributes
	err := gameobject.MergeAllAttributes(attributes, p.MaxStats, false)
	if err != nil {
		return err
	}
	for _, v := range equip {
		err := gameobject.MergeAllAttributes(attributes, v.Attributes, false)
		if err != nil {
			return err
		}
	}
	s, err := gameobject.SatisfyingAttributes(attributes, requirements)
	if err != nil {
		return err
	}
	if !s {
		return fmt.Errorf("item requirements are not satisfied")
	}

	return nil
}

func ExecuteYell(game *Game, p gameobject.Positioner, message *api.Message) error {
	messageEvent := api.Event_MESSAGE
	game.LogEvent(&api.Event{
		Type:        &messageEvent,
		Message:     fmt.Sprintf("%s (%s): %s", p.GetId(), p.GetName(), message.Text),
		Coordinates: p.GetPosition(),
	})
	return nil
}

func ExecuteBuy(game *Game, p *gameobject.Player, identifiers *api.Identifiers) error {
	err := ValidateBuy(game, p, identifiers)
	if err != nil {
		return err
	}

	for _, itemId := range identifiers.Ids {
		maybeItem, err := game.GetObjectById(itemId)
		if err != nil {
			return err
		}
		item, ok := maybeItem.(*api.Item)
		if !ok {
			return fmt.Errorf("%s is not Item ID", itemId)
		}
		if p.Character.Money < item.Price {
			// this should not happen (thanks to the validation above)
			return fmt.Errorf("insufficient funds to make the purchase")
		}
		p.Character.Money -= item.Price
		buyEvent := api.Event_BUY
		game.LogEvent(&api.Event{
			Type: &buyEvent,
			Message: fmt.Sprintf("Character %s (%s) bought item %s (%s)",
				p.Character.Id, p.Character.Name, itemId, item.Name)})

		// Buying also means equip in the version without inventory
		err = Equip(game, p, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func Equip(game *Game, player *gameobject.Player, item *api.Item) error {
	equipEvent := api.Event_EQUIP
	game.LogEvent(&api.Event{
		Type: &equipEvent,
		Message: fmt.Sprintf("Character %s (%s) equipped item %s (%s)",
			player.Character.Id, player.Character.Name, item.Id, item.Name)})
	return player.Equip(item)
}

func ExecutePickUp(game *Game, p *gameobject.Player, i *api.Identifier) error {
	// TODO solve concurrent pickUp (more than one player wants to pick up the same item)
	// TODO how to solve attributes consistently
	return fmt.Errorf("not implemented")
}

func payForSkill(p *gameobject.Player, s *api.Skill) error {
	if s.Cost == nil {
		return nil
	}
	return gameobject.SubtractAllAttributes(p.Character.Attributes, s.Cost, false)

}

func ExecuteSkill(game *Game, player *gameobject.Player, su *api.SkillUse) error {
	s, ok := player.Skills[su.SkillId]
	if !ok {
		return fmt.Errorf("skill %s not found for character", su.SkillId)
	}
	skillEvent := api.Event_SKILL
	game.LogEvent(
		&api.Event{
			Type:        &skillEvent,
			Message:     fmt.Sprintf("%s (%s): used skill: %s (%s)", player.Character.Id, player.Character.Name, s.Id, s.Name),
			Coordinates: player.Position,
		})

	err := payForSkill(player, s)
	if err != nil {
		return err
	}

	var duration float64
	if s.Duration != nil {
		d, err := gameobject.AttributesValue(player.Character.Attributes, s.Duration)
		if err != nil {
			return err
		}
		duration = gameobject.RoundSkill(d)
	}

	if s.CasterEffects != nil {
		e, err := gameobject.EvaluateSkillAttributes(s.CasterEffects.Attributes, player.Character.Attributes)
		if err != nil {
			return err
		}
		player.Character.Effects = append(player.Character.Effects, &api.Effect{
			Effects:   e,
			Duration:  int32(duration),
			XCasterId: player.GetId(),
		})
		// TODO summons
		// TODO flags
	}

	switch s.Target {
	case api.Skill_character:
		character, err := game.GetObjectById(*su.TargetId)
		if err != nil {
			return err
		}
		if s.TargetEffects != nil {
			e, err := gameobject.EvaluateSkillAttributes(s.TargetEffects.Attributes, player.Character.Attributes)
			if err != nil {
				return err
			}
			d, err := gameobject.AttributesValue(player.Character.Attributes, s.DamageAmount)
			if err != nil {
				return err
			}
			// TODO summons
			// TODO flags
			switch c := character.(type) {
			case *gameobject.Monster:
				c.Monster.Effects = append(c.Monster.Effects, &api.Effect{
					Effects:      e,
					DamageAmount: float32(d),
					DamageType:   s.DamageType,
					Duration:     int32(duration),
					XCasterId:    player.GetId(),
				})
			case *gameobject.Player:
				c.Character.Effects = append(c.Character.Effects, &api.Effect{
					Effects:      e,
					DamageAmount: float32(d),
					DamageType:   s.DamageType,
					Duration:     int32(duration),
					XCasterId:    player.GetId(),
				})
			default:
				return fmt.Errorf("tried to cast character spell on non-character")
			}
		}
	case api.Skill_item:
		//maybeItem, err := game.GetObjectById(*su.TargetId)
		//if err != nil {
		//	return err
		//}
		//i, ok := maybeItem.(*api.Item)
		//if !ok {
		//	return fmt.Errorf("tried to cast item spell on non-item")
		//}
		// TODO item effects?
	case api.Skill_position:
		// teleport
		err = game.MovePlayer(player, gameobject.PositionToCoordinates(su.Position, player.Position.Level))
		if err != nil {
			return err
		}
		return fmt.Errorf("not implemented")
	}
	return nil
}

func ExecuteAssignSkillPoints(player *gameobject.Player, a *api.Attributes) error {
	s, err := gameobject.SumAttributes(a)
	if err != nil {
		return err
	}
	player.Character.SkillPoints -= s
	err = gameobject.MergeAllAttributes(player.BaseAttributes, a, false)
	if err != nil {
		return err
	}
	player.ResetAttributes()
	return nil
}

func EvaluateEffects(g *Game, effects []*api.Effect, a *api.Attributes, receiver gameobject.Positioner) ([]*api.Effect, error) {
	var keptEffects []*api.Effect
	for _, e := range effects {
		err := gameobject.MergeAllAttributes(a, e.Effects, false)
		damage := gameobject.EvaluateDamage(float64(e.DamageAmount), e.DamageType, a)
		if err != nil {
			return keptEffects, err
		}

		attacker, err := g.GetObjectById(e.XCasterId)
		var attackerName string
		if err != nil {
			attackerName = attacker.GetName()
		}

		damageEvent := api.Event_DAMAGE
		g.LogEvent(&api.Event{
			Type:        &damageEvent,
			Message:     fmt.Sprintf("%s (%s): damaged %s (%s): %f with %s", e.XCasterId, attackerName, receiver.GetId(), receiver.GetName(), damage, e.DamageType.String()),
			Coordinates: receiver.GetPosition(),
		})

		e.Duration--
		if e.Duration > 0 {
			keptEffects = append(keptEffects, e)
		}
	}
	return keptEffects, nil
}
