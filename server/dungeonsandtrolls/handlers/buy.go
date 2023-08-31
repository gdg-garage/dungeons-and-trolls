package handlers

import (
	"fmt"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func validateIdentifiers(game *dungeonsandtrolls.Game, identifiers *api.Identifiers) error {
	for _, id := range identifiers.Ids {
		if _, ok := game.IdToItem[id]; !ok {
			return fmt.Errorf("item with ID %s does not exist", id)
		}
	}
	return nil
}

func Buy(game *dungeonsandtrolls.Game, identifiers *api.Identifiers) error {
	// TODO use apiKey
	p, err := game.GetPlayerByKey("test")
	if err != nil {
		return err
	}
	// TODO check player is located on the 0th floor
	// TODO lock
	err = validateIdentifiers(game, identifiers)
	if err != nil {
		return err
	}
	for _, itemId := range identifiers.Ids {
		item := game.IdToItem[itemId]
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
