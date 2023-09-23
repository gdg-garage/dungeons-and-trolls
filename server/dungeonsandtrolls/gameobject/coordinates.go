package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

func PositionToCoordinates(p *api.Position, l int32) *api.Coordinates {
	return &api.Coordinates{
		PositionX: p.PositionX,
		PositionY: p.PositionY,
		Level:     &l,
	}
}

func CoordinatesToPosition(coordinates *api.Coordinates) *api.Position {
	return &api.Position{
		PositionY: coordinates.PositionY,
		PositionX: coordinates.PositionX,
	}
}
