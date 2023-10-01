package gameobject

type MapMetadata struct {
	GeneratedTick      int32
	LastInteractedTick int32
	Fov                [][]bool
}

func IsMapDeprecated(mm *MapMetadata, t int32, l int32) bool {
	if l == 0 {
		if t-mm.GeneratedTick > 30 {
			return true
		}
	} else {
		if t-mm.GeneratedTick > 4*60 {
			return true
		}
		if t-mm.LastInteractedTick > 60 {
			return true
		}
	}
	return false
}
