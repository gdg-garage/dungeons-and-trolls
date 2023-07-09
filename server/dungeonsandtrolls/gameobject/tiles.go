package gameobject

func CreateEmpty() *GameObject {
	return &GameObject{
		Type: "Empty",
	}
}

func CreateWall() *GameObject {
	return &GameObject{
		Type: "Wall",
	}
}
