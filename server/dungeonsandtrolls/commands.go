package dungeonsandtrolls

type Direction string

const (
	Up    Direction = "up"
	Down  Direction = "down"
	Left  Direction = "left"
	Right Direction = "right"
)

type CommandI interface {
	GetType() string
}

type Command struct {
	Type string `json:"type"`
}

func (c Command) GetType() string {
	return c.Type
}

type MoveCommand struct {
	Command   `json:",inline"`
	Direction Direction `json:"direction"`
}

type AttackCommand struct {
	Command `json:",inline"`
	Target  string `json:"target"`
}

// Instants
type UseCommand struct {
	Command `json:",inline"`
	Target  string `json:"target"`
}
