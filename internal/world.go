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

type Rarity int

const (
	Junk Rarity = iota
	Common
	Uncommon
	Rare
	Epic
)

func (r Rarity) Of(msg string) string {
	switch r {
	case Junk:
		return WithColor(LightGray, msg)
	case Common:
		return WithColor(White, msg)
	case Uncommon:
		return WithColor(Green, msg)
	case Rare:
		return WithColor(Blue, msg)
	case Epic:
		return WithColor(Purple, msg)
	}
	return msg
}

func getRarity() Rarity {
	v := rand.Float32()
	if v > 0.95 {
		return Epic
	} else if v > 0.8 {
		return Rare
	} else if v > 0.5 {
		return Uncommon
	} else if v > 0.1 {
		return Common
	} else {
		return Junk
	}
}

type CellType int

const (
	CellTypeEmpty CellType = iota
	CellTypeDelta
	CellTypeOmega
	CellTypeSigma
	CellTypeLambda
)

func (ct CellType) String() string {
	switch ct {
	case CellTypeDelta:
		return DeltaSymbol
	case CellTypeOmega:
		return OmegaSymbol
	case CellTypeSigma:
		return SigmaSymbol
	case CellTypeLambda:
		return LambdaSymbol
	default:
		return " "
	}
}

type Cell struct {
	Type  CellType
	Value Rarity
}

func (c Cell) String() string {
	return c.Value.Of(c.Type.String())
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
			x := rand.Intn(level.Width - 1)
			y := rand.Intn(level.Height - 1)
			c := Coordinate{x, y}
			if grid[c].Type == CellTypeEmpty {
				grid[c] = Cell{cellType, getRarity()}
				filled++
			}
		}
	}

	return World{level.Width, level.Height, grid, bitStream}
}

func (w *World) shiftBitStream(x, y int) {
	newStream := make(map[Coordinate]string)
	// TODO: generate new bits
	for coord, val := range w.BitStream {
		c := Coordinate{
			X: (coord.X + x) % w.width,
			Y: (coord.Y + y) % w.height,
		}
		newStream[c] = val
	}
	w.BitStream = newStream
}
