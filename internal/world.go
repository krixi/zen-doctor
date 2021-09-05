package zen_doctor

import "math/rand"

type Coordinate struct {
	X int
	Y int
}

type Direction int

const (
	MoveUp Direction = iota
	MoveDown
	MoveLeft
	MoveRight
)

type CellType int

const (
	CellTypeEmpty CellType = iota
	CellTypeDelta
	CellTypeOmega
	CellTypeSigma
	CellTypeLambda
)

type Cell struct {
	Type CellType
}

type World struct {
	width     int
	height    int
	Grid      map[Coordinate]Cell
	BitStream map[Coordinate]string
}

var choices = []string{
	"0",
	"1",
}

func getBit() string {
	if rand.Float32() < 0.2 {
		return choices[rand.Intn(len(choices))]
	} else {
		return " "
	}
}

func newWorld(level LevelSettings) World {

	bitStream := make(map[Coordinate]string)
	grid := make(map[Coordinate]Cell)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			grid[c] = Cell{
				Type: CellTypeEmpty,
			}
			bitStream[c] = getBit()
		}
	}

	for cellType, count := range level.DataRequired {
		filled := 0
		for filled < count {
			// make sure it's empty first
			x := rand.Intn(level.Width)
			y := rand.Intn(level.Height)
			c := Coordinate{x, y}
			if grid[c].Type == CellTypeEmpty {
				grid[c] = Cell{cellType}
				filled++
			}
		}
	}

	return World{level.Width, level.Height, grid, bitStream}
}

func (w *World) shiftBitStream(x, y int) {
	newStream := make(map[Coordinate]string)
	for coord, val := range w.BitStream {
		c := Coordinate{
			X: (coord.X + x) % w.width,
			Y: (coord.Y + y) % w.height,
		}
		newStream[c] = val
	}
	w.BitStream = newStream
}
