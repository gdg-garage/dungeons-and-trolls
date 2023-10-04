package dungeonsandtrolls

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
	"go.openly.dev/pointy"
	"golang.org/x/exp/slices"
)

// ValidateBuy validates identifiers, funds, and requirements
func ValidateBuy(game *Game, p *gameobject.Player, identifiers *api.Identifiers) error {
	if p.Stun().IsStunned {
		return fmt.Errorf("you are stunned")
	}

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

func payForSkill(p gameobject.Skiller, s *api.Skill) error {
	if s.Cost == nil {
		return nil
	}
	return gameobject.SubtractAllAttributes(p.GetAttributes(), s.Cost, false)
}

func summon(game *Game, sum *api.Droppable, player gameobject.Skiller, duration int32) {
	po := game.GetMapObjectsOrCreateDefault(player.GetPosition())
	switch so := sum.Data.(type) {
	case *api.Droppable_Monster:
		so.Monster.Id = gameobject.GetNewId()
		if so.Monster.Faction == "inherited" {
			switch p := player.(type) {
			case *gameobject.Player:
				so.Monster.Faction = "player"
			case *gameobject.Monster:
				so.Monster.Faction = p.Monster.Faction
			}
		}
		po.Monsters = append(po.Monsters, so.Monster)
		moGo := gameobject.CreateMonster(so.Monster, player.GetPosition())
		moGo.KillCounter = &duration
		game.Register(moGo)
	case *api.Droppable_Decoration:
		po.Decorations = append(po.Decorations, so.Decoration)
	default:
		log.Warn().Msgf("summon of something unexpected attempted %s", sum.Data)
	}
}

func ExecuteSkill(game *Game, player gameobject.Skiller, su *api.SkillUse) error {
	s, ok := player.GetSkill(su.SkillId)
	if !ok {
		return fmt.Errorf("skill %s not found for character", su.SkillId)
	}
	skillEvent := api.Event_SKILL
	aoeEvent := api.Event_AOE

	event := &api.Event{
		Type:        &skillEvent,
		Message:     fmt.Sprintf("%s (%s): used skill: %s (%s)", player.GetId(), player.GetName(), s.Id, s.Name),
		SkillName:   &s.Name,
		PlayerId:    pointy.String(player.GetId()),
		Coordinates: player.GetPosition(),
		Skill:       s,
	}
	defer game.LogEvent(event)

	err := payForSkill(player, s)
	if err != nil {
		return err
	}

	var duration float64
	if s.Duration != nil {
		d, err := gameobject.AttributesValue(player.GetAttributes(), s.Duration)
		if err != nil {
			return err
		}
		duration = gameobject.RoundSkill(d)
	}

	radiusValue, err := gameobject.AttributesValue(player.GetAttributes(), s.Radius)
	if err != nil {
		return err
	}
	radiusValue = gameobject.RoundSkill(radiusValue)

	var targetPos *api.Coordinates
	switch s.Target {
	case api.Skill_character:
		character, err := game.GetObjectById(*su.TargetId)
		if err != nil {
			return err
		}
		switch ch := character.(type) {
		case *gameobject.Player:
			targetPos = ch.GetPosition()
		case *gameobject.Monster:
			targetPos = ch.GetPosition()
		default:
			return fmt.Errorf("targetPos is not a monster or player")
		}
	case api.Skill_position:
		targetPos = gameobject.PositionToCoordinates(su.Position, player.GetPosition().Level)
	case api.Skill_none:
		targetPos = player.GetPosition()
	}

	event.Target = targetPos

	if s.CasterEffects != nil {
		e, err := gameobject.EvaluateSkillAttributes(s.CasterEffects.Attributes, player.GetAttributes())
		if err != nil {
			return err
		}
		switch pt := player.(type) {
		case gameobject.Skiller:
			casterId := player.GetId()
			if s.CasterEffects.Flags.Stun {
				pt.Stunned()
			}
			pt.AddEffect(&api.Effect{
				Effects:   e,
				Duration:  int32(duration),
				XCasterId: &casterId,
			})

			if s.CasterEffects.Flags.Movement {
				gameobject.TeleportMoveTo(pt, targetPos)
			}
		}

		for _, sum := range s.CasterEffects.Summons {
			summon(game, sum, player, int32(duration))
		}

		if radiusValue > 0 {
			game.LogEvent(&api.Event{
				Type:        &aoeEvent,
				Message:     fmt.Sprintf("%s (%s): caused aoe effect", player.GetId(), player.GetName()),
				SkillName:   &s.Name,
				Coordinates: player.GetPosition(),
				Radius:      pointy.Float32(float32(radiusValue)),
			})
			//TODO
			//for _, t := range handlers.TilesInRange(game, player.GetPosition(), int32(radiusValue)) {
			//	//for t.Monsters
			//	//for t.Players
			//}
		}

		// TODO AOE
		// TODO flags (knockback, ground)
	}

	switch s.Target {
	case api.Skill_character:
		character, err := game.GetObjectById(*su.TargetId)
		if err != nil {
			return err
		}
		if s.TargetEffects != nil {
			e, err := gameobject.EvaluateSkillAttributes(s.TargetEffects.Attributes, player.GetAttributes())
			if err != nil {
				return err
			}
			d, err := gameobject.AttributesValue(player.GetAttributes(), s.DamageAmount)
			if err != nil {
				return err
			}

			switch c := character.(type) {
			case gameobject.Skiller:
				casterId := player.GetId()
				c.Stunned()
				c.AddEffect(&api.Effect{
					Effects:      e,
					DamageAmount: float32(d),
					DamageType:   s.DamageType,
					Duration:     int32(duration),
					XCasterId:    &casterId,
				})
				for _, sum := range s.CasterEffects.Summons {
					summon(game, sum, c, int32(duration))
				}

				if s.CasterEffects.Flags.Movement {
					gameobject.TeleportMoveTo(c, targetPos)
				}

				if radiusValue > 0 {
					game.LogEvent(&api.Event{
						Type:        &aoeEvent,
						Message:     fmt.Sprintf("%s (%s): caused aoe effect", player.GetId(), player.GetName()),
						SkillName:   &s.Name,
						Coordinates: player.GetPosition(),
						Radius:      pointy.Float32(float32(radiusValue)),
					})
					//TODO
					//for _, t := range handlers.TilesInRange(game, player.GetPosition(), int32(radiusValue)) {
					//	//for t.Monsters
					//	//for t.Players
					//}
				}

				// TODO flags (knockback)
			default:
				return fmt.Errorf("tried to cast character spell on non-character")
			}
		}
	case api.Skill_position:
		//return fmt.Errorf("not implemented")
	}

	return nil
}

func ExecuteAssignSkillPoints(player *gameobject.Player, a *api.Attributes) error {
	s, err := gameobject.SumAttributes(a)
	if err != nil {
		return err
	}
	log.Info().Msgf("player %s (%s) is assigning attributes %+v", player.GetId(), player.GetName(), a)
	player.Character.SkillPoints -= s
	err = gameobject.MergeAllAttributes(player.BaseAttributes, a, false)
	if err != nil {
		return err
	}
	player.UpdateAttributes()
	return nil
}

func EvaluateEffects(g *Game, effects []*api.Effect, a *api.Attributes, receiver gameobject.Alive) ([]*api.Effect, error) {
	var keptEffects []*api.Effect
	var errIdx []int

	// Apply buffs - maybe it will grant more resist
	for i, e := range effects {
		err := gameobject.MergeAllAttributes(a, e.Effects, false)
		if err != nil {
			errIdx = append(errIdx, i)
		}
	}

	// Deal damage
	for _, e := range effects {
		if e.DamageType != api.DamageType_none {
			damage := gameobject.EvaluateDamage(float64(e.DamageAmount), e.DamageType, a)

			var attackerName string
			if e.XCasterId != nil {
				attacker, err := g.GetObjectById(*e.XCasterId)
				if err == nil {
					attackerName = attacker.GetName()
				}
			}

			if damage > 0 {
				receiver.DamageTaken()
			}

			damageEvent := api.Event_DAMAGE
			g.LogEvent(&api.Event{
				Type:        &damageEvent,
				Message:     fmt.Sprintf("%s (%s): damaged %s (%s): %f with %s", *e.XCasterId, attackerName, receiver.GetId(), receiver.GetName(), damage, e.DamageType.String()),
				Damage:      &damage,
				Coordinates: receiver.GetPosition(),
			})
		}
	}

	// what to keep
	for i, e := range effects {
		// skip effects which caused err
		if slices.Contains(errIdx, i) {
			continue
		}

		e.Duration--
		if e.Duration > 0 {
			keptEffects = append(keptEffects, e)
		}
	}
	return keptEffects, nil
}
