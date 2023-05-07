package gameobject

type GameObject struct {
	Type     string      `json:"type"`
	Children []Interface `json:"children"`
}

type Interface interface {
	GetType() string
}

func (g *GameObject) GetType() string {
	return g.Type
}
