package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

func FilterEffect(e *api.Effect) {
	e.XCasterId = nil
}
