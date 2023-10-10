package dungeonsandtrolls

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"github.com/rs/zerolog/log"
	"go.openly.dev/pointy"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
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
	err := gameobject.MergeAllAttributes(attributes, p.BaseAttributes, false)
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
	// Check that object is still in the game - it is unregistered on pickup (first player takes the item).
	// it should not be common - it may be improved.
	maybeItem, err := game.GetObjectById(i.Id)
	item, ok := maybeItem.(*api.Item)
	if !ok {
		log.Warn().Msgf("What should be an item is not")
	}
	if err != nil {
		log.Warn().Err(err).Msgf("tried to pick up non-existent item")
		return nil
	}
	err = p.Equip(item)
	if err != nil {
		log.Warn().Err(err).Msgf("equip of the picked up item %s failed", i.GetId())
	}
	err = p.UpdateAttributes()
	if err != nil {
		log.Warn().Err(err).Msgf("player attributes update failed after pick up")
	}
	o, err := game.GetObjectsOnPosition(p.GetPosition())
	if err != nil {
		log.Warn().Err(err).Msgf("position of picked up item %s is malformed", i.GetId())
	}
	game.removeItemFromTile(o, item)
	return nil
}

func payForSkill(p gameobject.Skiller, s *api.Skill) error {
	if s.Cost == nil {
		return nil
	}
	return gameobject.SubtractAllAttributes(p.GetAttributes(), s.Cost, false)
}

func summon(game *Game, sum *api.Droppable, player gameobject.Skiller, target *api.Coordinates, duration int32) {
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
		moGo := gameobject.CreateMonster(so.Monster, target)
		moGo.KillCounter = &duration
		game.Register(moGo)
	case *api.Droppable_Decoration:
		po.Decorations = append(po.Decorations, so.Decoration)
	default:
		log.Warn().Msgf("summon of something unexpected attempted %s", sum.Data)
	}
}

func TilesInRange(game *Game, startingPosition *api.Coordinates, rng int32) []*api.MapObjects {
	var aoe []*api.MapObjects
	lc, err := game.GetCachedLevel(startingPosition.Level)
	if err != nil {
		log.Warn().Err(err).Msgf("level retrieval failed for aoe")
		return aoe
	}
	p := proto.Clone(startingPosition).(*api.Coordinates)
	for x := startingPosition.PositionX - rng; x < startingPosition.PositionX+rng; x++ {
		if x < 0 || x >= lc.Height {
			continue
		}
		p.PositionX = x
		for y := startingPosition.PositionY - rng; y < startingPosition.PositionY+rng; y++ {
			if y < 0 || y >= lc.Width {
				continue
			}
			p.PositionY = y
			dist := utils.ManhattanDistance(startingPosition.PositionX, startingPosition.PositionY, p.PositionX, p.PositionY)
			if dist > rng {
				continue
			}
			mo, err := game.GetObjectsOnPosition(p)
			if mo == nil {
				log.Warn().Err(err).Msgf("tile retrieval failed for aoe")
				continue
			}
			aoe = append(aoe, mo)
		}
	}

	return aoe
}

func findSkillerAndAddEffect(g *Game, tid string, damage float64, s *api.Skill, e *api.Attributes, duration int32, casterId string, targetPos *api.Coordinates) error {
	soi, err := g.GetObjectById(tid)
	so := soi.(gameobject.Skiller)
	if err != nil {
		return err
	}

	so.AddEffect(&api.Effect{
		Effects:      e,
		DamageAmount: float32(damage),
		DamageType:   s.DamageType,
		Duration:     duration,
		XCasterId:    &casterId,
	})

	if s.CasterEffects.Flags.Stun {
		so.Stunned()
	}

	if s.CasterEffects.Flags.Movement {
		gameobject.TeleportMoveTo(so, targetPos)
	}

	if s.CasterEffects.Flags.Knockback {
		// todo knockback
	}

	return nil
}

func ExecuteSkill(game *Game, player gameobject.Skiller, su *api.SkillUse) error {
	s, ok := player.GetSkill(su.SkillId)
	if !ok {
		return fmt.Errorf("skill %s not found for character", su.SkillId)
	}
	skillEvent := api.Event_SKILL
	aoeEvent := api.Event_AOE

	distanceValue, err := gameobject.AttributesValue(player.GetAttributes(), s.Range)
	if err != nil {
		return err
	}
	distanceValue = gameobject.RoundRange(distanceValue)
	ranged := false
	if distanceValue > 3 {
		ranged = true
	}

	event := &api.Event{
		Type:        &skillEvent,
		Message:     fmt.Sprintf("%s (%s): used skill: %s (%s)", player.GetId(), player.GetName(), s.Id, s.Name),
		SkillName:   &s.Name,
		PlayerId:    pointy.String(player.GetId()),
		Coordinates: player.GetPosition(),
		Skill:       s,
		IsRanged:    &ranged,
	}
	defer game.LogEvent(event)

	err = payForSkill(player, s)
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

	d, err := gameobject.AttributesValue(player.GetAttributes(), s.DamageAmount)
	if err != nil {
		return err
	}

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

	casterId := player.GetId()

	if s.CasterEffects != nil {
		e, err := gameobject.EvaluateSkillAttributes(s.CasterEffects.Attributes, player.GetAttributes())
		if err != nil {
			return err
		}
		switch pt := player.(type) {
		case gameobject.Skiller:
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
			summon(game, sum, player, player.GetPosition(), int32(duration))
		}

		if s.CasterEffects.Flags.Knockback {
			// TODO flags knockback
		}
	}

	if s.CasterEffects.Flags.GroundEffect {
		game.LogEvent(&api.Event{
			Type:        &aoeEvent,
			Message:     fmt.Sprintf("%s (%s): caused aoe ground effect", player.GetId(), player.GetName()),
			SkillName:   &s.Name,
			Coordinates: player.GetPosition(),
			Radius:      pointy.Float32(float32(radiusValue)),
		})

		e, err := gameobject.EvaluateSkillAttributes(s.CasterEffects.Attributes, player.GetAttributes())
		if err != nil {
			return err
		}

		for _, t := range TilesInRange(game, player.GetPosition(), int32(radiusValue)) {
			t.Effects = append(t.Effects, &api.Effect{
				Effects:      e,
				DamageAmount: float32(d),
				DamageType:   s.DamageType,
				Duration:     int32(duration),
				XCasterId:    &casterId,
			})

			// stun should be applied multiple times (but it is not supported now)
			if s.CasterEffects.Flags.Stun {
				for _, p := range t.Players {
					soi, err := game.GetObjectById(p.GetId())
					so := soi.(gameobject.Skiller)
					if err != nil {
						return err
					}

					if s.CasterEffects.Flags.Stun {
						so.Stunned()
					}
				}

				for _, p := range t.Monsters {
					soi, err := game.GetObjectById(p.GetId())
					so := soi.(gameobject.Skiller)
					if err != nil {
						return err
					}

					if s.CasterEffects.Flags.Stun {
						so.Stunned()
					}
				}
			}
		}
	}

	if s.TargetEffects.Flags.GroundEffect {
		game.LogEvent(&api.Event{
			Type:        &aoeEvent,
			Message:     fmt.Sprintf("%s (%s): caused aoe ground effect", player.GetId(), player.GetName()),
			SkillName:   &s.Name,
			Coordinates: targetPos,
			Radius:      pointy.Float32(float32(radiusValue)),
		})

		e, err := gameobject.EvaluateSkillAttributes(s.TargetEffects.Attributes, player.GetAttributes())
		if err != nil {
			return err
		}

		for _, t := range TilesInRange(game, targetPos, int32(radiusValue)) {
			t.Effects = append(t.Effects, &api.Effect{
				Effects:      e,
				DamageAmount: float32(d),
				DamageType:   s.DamageType,
				Duration:     int32(duration),
				XCasterId:    &casterId,
			})

			// stun should be applied multiple times (but it is not supported now)
			if s.CasterEffects.Flags.Stun {
				for _, p := range t.Players {
					soi, err := game.GetObjectById(p.GetId())
					so := soi.(gameobject.Skiller)
					if err != nil {
						return err
					}

					if s.CasterEffects.Flags.Stun {
						so.Stunned()
					}
				}

				for _, p := range t.Monsters {
					soi, err := game.GetObjectById(p.GetId())
					so := soi.(gameobject.Skiller)
					if err != nil {
						return err
					}

					if s.CasterEffects.Flags.Stun {
						so.Stunned()
					}
				}
			}
		}
	}

	if s.CasterEffects.Flags.GroundEffect || s.TargetEffects.Flags.GroundEffect {
		return nil
	}

	if s.TargetEffects != nil {
		e, err := gameobject.EvaluateSkillAttributes(s.TargetEffects.Attributes, player.GetAttributes())
		if err != nil {
			return err
		}

		// AOE
		if radiusValue > 0 {
			game.LogEvent(&api.Event{
				Type:        &aoeEvent,
				Message:     fmt.Sprintf("%s (%s): caused aoe effect", player.GetId(), player.GetName()),
				SkillName:   &s.Name,
				Coordinates: targetPos,
				Radius:      pointy.Float32(float32(radiusValue)),
			})
			for _, t := range TilesInRange(game, targetPos, int32(radiusValue)) {
				for _, m := range t.Monsters {
					err := findSkillerAndAddEffect(game, m.GetId(), d, s, e, int32(duration), casterId, targetPos)
					if err != nil {
						return err
					}
				}
				for _, p := range t.Players {
					if p.GetId() == player.GetId() && s.Target != api.Skill_character {
						continue
					}
					err := findSkillerAndAddEffect(game, p.GetId(), d, s, e, int32(duration), casterId, targetPos)
					if err != nil {
						return err
					}
				}
			}
		} else if s.Target == api.Skill_position {
			// apply to everyone (else) on the same tile
			for _, t := range TilesInRange(game, targetPos, 0) {
				for _, m := range t.Monsters {
					err := findSkillerAndAddEffect(game, m.GetId(), d, s, e, int32(duration), casterId, targetPos)
					if err != nil {
						return err
					}
				}
				for _, p := range t.Players {
					if p.GetId() == player.GetId() {
						continue
					}
					err := findSkillerAndAddEffect(game, p.GetId(), d, s, e, int32(duration), casterId, targetPos)
					if err != nil {
						return err
					}
				}
			}
		} else if s.Target == api.Skill_character {
			// apply to the target char regardless the range
			err := findSkillerAndAddEffect(game, *su.TargetId, d, s, e, int32(duration), casterId, targetPos)
			if err != nil {
				return err
			}
		}

		// TODO flags (knockback)

		for _, sum := range s.TargetEffects.Summons {
			summon(game, sum, player, targetPos, int32(duration))
		}
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
