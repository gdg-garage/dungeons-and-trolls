package handlers

import (
	"fmt"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

// validate identifiers, funds, and requirements
func validateBuy(game *dungeonsandtrolls.Game, identifiers *api.Identifiers, p *gameobject.Player) error {
	// this has to be a copy
	playerMoney := p.Character.Money
	for _, id := range identifiers.Ids {
		maybeItem, err := game.GetObjectById(id)
		if err != nil {
			return err
		}
		i, ok := maybeItem.(*api.Item)
		if !ok {
			return fmt.Errorf("%s is not an Item ID", id)
		}
		playerMoney -= i.BuyPrice
		if playerMoney < 0 {
			return fmt.Errorf("insufficient funds to make the purchase")
		}
	}
	// TODO check requirements
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

	err = validateBuy(game, identifiers, p)
	if err != nil {
		return err
	}

	pc := game.GetPlayerCommands(p.Character.Id)
	pc.Buy = identifiers

	return nil
}
