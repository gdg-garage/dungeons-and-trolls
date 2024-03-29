package handlers

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"testing"
)

func TestRegistration(t *testing.T) {
	game := dungeonsandtrolls.NewGame()
	game.Players["player 1"] = gameobject.CreatePlayer("player 1")
	if validateRegistration(game, "") == nil {
		t.Fatal("empty user allowed")
	}
	if validateRegistration(game, "player 1") == nil {
		t.Fatal("existing user allowed")
	}
	if validateRegistration(game, "player 2") != nil {
		t.Fatal("registration failed")
	}
}
