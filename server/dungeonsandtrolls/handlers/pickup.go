package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"google.golang.org/protobuf/proto"
)

func validatePickUp(game *dungeonsandtrolls.Game, i *api.Identifier, p *gameobject.Player) error {
	// TODO maybe buyValidation could be used if the price is 0?
	// check item (exists and is item)
	o, err := game.GetObjectsOnPosition(p.Position)
	if err != nil {
		return err
	}
	if o == nil {
		return fmt.Errorf("there are no objects in the pickup location (player posistion)")
	}
	var item *api.Item
	for _, it := range o.Items {
		if it.Id == i.Id {
			item = it
		}
	}
	if item == nil {
		return fmt.Errorf("there are is no item %s in pickup location (player posistion)", i.Id)
	}

	// check requirements
	attributes := proto.Clone(p.Character.Attributes).(*api.Attributes)
	e, ok := p.Equipped[item.Slot]
	if ok {
		err := gameobject.SubtractAllAttributes(attributes, e.Attributes, true)
		if err != nil {
			return err
		}
	}
	err = gameobject.MergeAllAttributes(attributes, item.Attributes, false)
	if err != nil {
		return err
	}
	s, err := gameobject.SatisfyingAttributes(attributes, item.Requirements)
	if err != nil {
		return err
	}
	if !s {
		return fmt.Errorf("requirements not satisfied")
	}

	// check that requirements for all other items are still satisfied after the swap
	for socket, ei := range p.Equipped {
		if socket == item.Slot {
			continue
		}
		s, err := gameobject.SatisfyingAttributes(attributes, ei.Requirements)
		if err != nil {
			return err
		}
		if !s {
			return fmt.Errorf("requirements not satisfied")
		}
	}

	return nil
}

func PickUp(game *dungeonsandtrolls.Game, i *api.Identifier, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	err = validatePickUp(game, i, p)
	if err != nil {
		return err
	}

	pc := game.GetCommands(p.Character.Id)
	pc.PickUp = i

	return nil
}
