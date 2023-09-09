package handlers

import (
	"fmt"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

func validateIdentifiers(game *dungeonsandtrolls.Game, identifiers *api.Identifiers) error {
	for _, id := range identifiers.Ids {
		maybeItem, err := game.GetObjectById(id)
		if err != nil {
			return err
		}
		_, ok := maybeItem.(*api.Item)
		if !ok {
			return fmt.Errorf("%s is not Item ID", id)
		}
	}
	return nil
}

func Buy(game *dungeonsandtrolls.Game, identifiers *api.Identifiers, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}

	if p.Position.Level == &gameobject.ZeroLevel {
		return fmt.Errorf("buying is availible only on the ground floor")
	}

	err = validateIdentifiers(game, identifiers)
	if err != nil {
		return err
	}

	// TODO this should be deferred - not executed now
	// TODO lock

	for _, itemId := range identifiers.Ids {
		maybeItem, err := game.GetObjectById(itemId)
		if err != nil {
			return err
		}
		item, ok := maybeItem.(*api.Item)
		if !ok {
			return fmt.Errorf("%s is not Item ID", itemId)
		}
		p.Character.Money -= item.BuyPrice
		buyEvent := api.Event_BUY
		game.LogEvent(&api.Event{
			Type: &buyEvent,
			Message: fmt.Sprintf("Character %s (%s) bought item %s (%s)",
				p.Character.Id, p.Character.Name, itemId, item.Name)})
		// Buying also means equip in the version without inventory
		Equip(game, item, p)
	}
	return nil
}
