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
	Legendary
)

func getRarity() Rarity {
	v := rand.Float32()
	// TODO: make this loot table dynamic based on level difficulty
	if v > 0.98 {
		return Legendary
	} else if v > 0.93 {
		return Epic
	} else if v > 0.75 {
		return Rare
	} else if v > 0.5 {
		return Uncommon
	} else if v > 0.2 {
		return Common
	} else {
		return Junk
	}
}

type LootType int

const (
	LootTypeEmpty LootType = iota
	LootTypeDelta
	LootTypeOmega
	LootTypeSigma
	LootTypeLambda
)

func (lt LootType) String() string {
	switch lt {
	case LootTypeDelta:
		return DeltaSymbol
	case LootTypeOmega:
		return OmegaSymbol
	case LootTypeSigma:
		return SigmaSymbol
	case LootTypeLambda:
		return LambdaSymbol
	default:
		return " "
	}
}

type Loot struct {
	Type  LootType
	Value Rarity
}

func (d Loot) String() string {
	msg := d.Type.String()
	switch d.Value {
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
	case Legendary:
		return WithColor(Orange, msg)
	}
	return msg
}

type HiddenBitType int

const (
	BitTypeEmpty HiddenBitType = iota
	BitTypeZero
	BitTypeOne
)

func (b HiddenBitType) String() string {
	switch b {
	case BitTypeZero:
		return `0`
	case BitTypeOne:
		return `1`
	}
	return ` `
}

type RevealedBitType int

const (
	RevealedBitBenign RevealedBitType = iota
	RevealedBitHelpful
	RevealedBitHarmful
)

type Bits struct {
	Hidden   HiddenBitType
	Revealed RevealedBitType
	Value    Rarity
	Symbol   string
}

func (b Bits) ViewHidden() string {
	return b.Hidden.String()
}
func (b Bits) ViewRevealed() string {
	return b.Symbol
}

// Threat returns the magnitude of the threat based on the value.
// you still need to multiply by -1 if it's helpful
func (b Bits) Threat() float32 {
	// TODO: make this configurable based on the level
	switch b.Value {
	case Junk:
		return 1
	case Common:
		return 2
	case Uncommon:
		return 3
	case Rare:
		return 5
	case Epic:
		return 8
	case Legendary:
		return 15
	}
	return 0
}

var hiddenBits = []HiddenBitType{
	BitTypeZero,
	BitTypeOne,
}
var helpfulBits = []string{
	ShrugSymbol,
	PhiSymbol,
	DiamondSymbol,
}
var harmfulBits = []string{
	DaggerSymbol,
	KoppaSymbol,
	PsiSymbol,
}

// TODO: make this configurable based on level settings.
func getBit() Bits {
	hidden := BitTypeEmpty
	revealed := RevealedBitBenign
	rarity := Junk
	symbol := " "
	if rand.Float32() < 0.2 {
		hidden = hiddenBits[rand.Intn(len(hiddenBits))]
		symbol = hidden.String()
		rarity = getRarity()
		next := rand.Float32()
		if next < 0.1 {
			revealed = RevealedBitHarmful
			symbol = harmfulBits[rand.Intn(len(harmfulBits))]
		} else if next > 0.98 {
			revealed = RevealedBitHelpful
			symbol = helpfulBits[rand.Intn(len(helpfulBits))]
		}
	}
	return Bits{hidden, revealed, rarity, symbol}
}

type World struct {
	width     int
	height    int
	Loot      map[Coordinate]Loot
	BitStream map[Coordinate]Bits
}

func newWorld(level LevelSettings) World {

	bitStream := make(map[Coordinate]Bits)
	loot := make(map[Coordinate]Loot)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			loot[c] = Loot{
				Type: LootTypeEmpty,
			}
			bitStream[c] = getBit()
		}
	}

	for dataType, count := range level.DataRequired {
		filled := 0
		for filled < count {
			// make sure it's empty first
			x := rand.Intn(level.Width - 1)
			y := rand.Intn(level.Height - 1)
			c := Coordinate{x, y}
			if loot[c].Type == LootTypeEmpty {
				loot[c] = Loot{dataType, getRarity()}
				filled++
			}
		}
	}

	return World{level.Width, level.Height, loot, bitStream}
}

func (w *World) shiftBitStream(dir Direction) {
	newStream := make(map[Coordinate]Bits)

	switch dir {
	case MoveDown:
		// update coordinates for all existing items in the stream
		for coord, val := range w.BitStream {
			c := Coordinate{
				X: coord.X,
				Y: coord.Y + 1,
			}
			if c.Y < w.height {
				newStream[c] = val
			}
		}
		// Add a new row on top
		for x := 0; x < w.width; x++ {
			c := Coordinate{
				X: x,
				Y: 0,
			}
			newStream[c] = getBit()
		}
	}
	w.BitStream = newStream
}

func (w *World) DidCollideWith(c Coordinate, bitType RevealedBitType) (float32, bool) {
	if b, ok := w.BitStream[c]; ok {
		return b.Threat(), b.Revealed == bitType
	}
	return 0, false
}
