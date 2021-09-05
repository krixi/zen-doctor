package zen_doctor

import (
	"math/rand"
	"strings"
)

type Coordinate struct {
	X int
	Y int
}



type GameState struct {
	CurrentLevel Level
	View         map[Coordinate]string
}

func NewGameState() GameState {
	return GameState{
		CurrentLevel: Tutorial,
	}
}

var choices = []string{
	"0",
	"1",
}

func getCell() string {
	if rand.Float32() < 0.2 {
		return choices[rand.Intn(len(choices))]
	} else {
		return " "
	}
}

func (s *GameState) InitView() {
	level := GetLevel(s.CurrentLevel)
	s.View = make(map[Coordinate]string)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			s.View[c] = getCell()
		}
	}
}

func (s *GameState) String() string {
	level := GetLevel(s.CurrentLevel)
	b := strings.Builder{}
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			c := Coordinate{x, y}
			b.WriteString(s.View[c])
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (s *GameState) Shift(x, y int) {
	level := GetLevel(s.CurrentLevel)
	newView := make(map[Coordinate]string)
	for coord, val := range s.View {
		c := Coordinate{
			X: (coord.X + x) % level.Width,
			Y: (coord.Y + y) % level.Height,
		}
		newView[c] = val
	}
	s.View = newView
}
