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
	// TODO lock
	err = validateIdentifiers(game, identifiers)
	if err != nil {
		return err
	}
	for _, itemId := range identifiers.Ids {
		item := game.IdToItem[itemId]
		p.Character.Money -= item.Price
		// Buying also means equip in the version without inventory
		p.Equip(item)
	}
	return nil
}
