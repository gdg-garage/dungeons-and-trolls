package dungeonsandtrolls

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

type CommandI interface {
	GetType() string
}

type Command struct {
	Type string `json:"type"`
}

func (c Command) GetType() string {
	return c.Type
}

type MoveCommand struct {
	Command   `json:",inline"`
	Direction Direction `json:"direction"`
}

type AttackCommand struct {
	Command `json:",inline"`
	Target  string `json:"target"`
}

// Instants
type UseCommand struct {
	Command `json:",inline"`
	Target  string `json:"target"`
}

// ValidateBuy validates identifiers, funds, and requirements
func ValidateBuy(game *Game, identifiers *api.Identifiers, p *gameobject.Player) error {
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
		playerMoney -= i.BuyPrice
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

func ExecuteYell(game *Game, p *gameobject.Player, message *api.Message) error {
	messageEvent := api.Event_MESSAGE
	game.LogEvent(&api.Event{
		Type:        &messageEvent,
		Message:     fmt.Sprintf("%s (%s): %s", p.Character.Id, p.Character.Name, message.Text),
		Coordinates: p.Position,
	})
	return nil
}

func ExecuteBuy(game *Game, p *gameobject.Player, identifiers *api.Identifiers) error {
	err := ValidateBuy(game, identifiers, p)
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
		if p.Character.Money < item.BuyPrice {
			// this should not happen (thanks to the validation above)
			return fmt.Errorf("insufficient funds to make the purchase")
		}
		p.Character.Money -= item.BuyPrice
		buyEvent := api.Event_BUY
		game.LogEvent(&api.Event{
			Type: &buyEvent,
			Message: fmt.Sprintf("Character %s (%s) bought item %s (%s)",
				p.Character.Id, p.Character.Name, itemId, item.Name)})

		// Buying also means equip in the version without inventory
		err = Equip(game, item, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func Equip(game *Game, item *api.Item, player *gameobject.Player) error {
	equipEvent := api.Event_EQUIP
	game.LogEvent(&api.Event{
		Type: &equipEvent,
		Message: fmt.Sprintf("Character %s (%s) equipped item %s (%s)",
			player.Character.Id, player.Character.Name, item.Id, item.Name)})
	return player.Equip(item)
}
