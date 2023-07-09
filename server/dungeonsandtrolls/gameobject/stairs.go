package gameobject

type Stairs struct {
	GameObject `json:",inline"`
	LeadsTo    int `json:"leads-to"`
}

func CreateStairs(leadsToLevel int) *Stairs {
	return &Stairs{
		LeadsTo: leadsToLevel,
		GameObject: GameObject{
			Type: "Stairs",
		},
	}
}
