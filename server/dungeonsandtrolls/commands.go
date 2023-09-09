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
	// TODO validate requirements.

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
		if p.Character.Money < 0 {
			return fmt.Errorf("insufficient funds to make the purchase")
		}
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

func Equip(game *Game, item *api.Item, player *gameobject.Player) {
	equipEvent := api.Event_EQUIP
	game.LogEvent(&api.Event{
		Type: &equipEvent,
		Message: fmt.Sprintf("Character %s (%s) equipped item %s (%s)",
			player.Character.Id, player.Character.Name, item.Id, item.Name)})
	player.Equip(item)
}
