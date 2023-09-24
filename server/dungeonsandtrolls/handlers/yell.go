package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

func validateYell(game *dungeonsandtrolls.Game, message *api.Message) error {
	if len(message.Text) > 80 {
		return fmt.Errorf("message is too long >80 (we are not Twitter)")
	}
	return nil
}

func Yell(game *dungeonsandtrolls.Game, message *api.Message, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}
	if p.IsAdmin {
		return fmt.Errorf("admin players are are not allowed to call non-monster commands")
	}

	// TODO translate IDs to names
	// - consider IDs as one char?

	err = validateYell(game, message)
	if err != nil {
		return err
	}

	pc := game.GetCommands(p.Character.Id)
	pc.Yell = message

	return nil
}
