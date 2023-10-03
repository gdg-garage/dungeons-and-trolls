package handlers

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
)

func Respawn(game *dungeonsandtrolls.Game, token string) error {
	p, err := game.GetCurrentPlayer(token)
	if err != nil {
		return err
	}
	if p.IsAdmin {
		return fmt.Errorf("admin players are are not allowed to call non-monster commands")
	}

	// TODO defer this
	game.GameLock.Lock()
	game.Respawn(p, true)
	game.GameLock.Unlock()

	return nil
}
