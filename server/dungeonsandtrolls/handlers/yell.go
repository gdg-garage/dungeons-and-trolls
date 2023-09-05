package handlers

import (
	"fmt"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func Yell(game *dungeonsandtrolls.Game, message *api.Message) error {
	p, err := game.GetCurrentPlayer()
	if err != nil {
		return err
	}

	// TODO translate IDs to names
	// - consider IDs as one char?

	if len(message.Text) > 80 {
		return fmt.Errorf("message is too long >80 (we are not Twitter)")
	}

	messageEvent := api.Event_MESSAGE
	game.LogEvent(&api.Event{
		Type:        &messageEvent,
		Message:     fmt.Sprintf("%s (%s): %s", p.Character.Id, p.Character.Name, message.Text),
		Coordinates: p.Position,
	})
	return nil
}
