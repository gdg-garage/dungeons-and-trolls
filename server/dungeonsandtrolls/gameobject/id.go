package gameobject

import (
	"github.com/google/uuid"
)

type Ider interface {
	GetId() string
	GetName() string
	// TODO may contain name method too
}

func GetNewId() string {
	return uuid.New().String()
}
