package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

func HideUnidentifiedFields(i *api.Item) {
	i.Attributes = &api.Attributes{}
	i.Requirements = &api.Attributes{}
	i.Skills = []*api.Skill{}
}
