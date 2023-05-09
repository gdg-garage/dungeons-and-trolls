package gameobject

type GameObject struct {
	Type     string      `json:"type"`
	Children []Interface `json:"children"`
}

type Interface interface {
	GetType() string
	GetChildren() []Interface
	SetChildren([]Interface)
}

func (g *GameObject) GetType() string {
	return g.Type
}

func (g *GameObject) GetChildren() []Interface {
	return g.Children
}

func (g *GameObject) SetChildren(children []Interface) {
	g.Children = children
}
