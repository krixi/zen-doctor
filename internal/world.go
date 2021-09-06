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

func getRarity(level LevelConfig) Rarity {
	v := rand.Float32()
	if v > (1 - level.LootChanceByRarity[Legendary]) {
		return Legendary
	} else if v > (1 - level.LootChanceByRarity[Epic]) {
		return Epic
	} else if v > (1 - level.LootChanceByRarity[Rare]) {
		return Rare
	} else if v > (1 - level.LootChanceByRarity[Uncommon]) {
		return Uncommon
	} else if v > (1 - level.LootChanceByRarity[Common]) {
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
func (b Bits) Threat(level LevelConfig) float32 {
	if threat, ok := level.ThreatByRarity[b.Value]; ok {
		return threat
	}
	return 0
}

var hiddenBits = []HiddenBitType{
	BitTypeZero,
	BitTypeOne,
}
var helpfulBits = []string{
	WeirdFRuneSymbol,
	PhiSymbol,
	DiamondSymbol,
}
var harmfulBits = []string{
	DaggerSymbol,
	KoppaSymbol,
	PsiSymbol,
}

func getBit(level LevelConfig) Bits {
	hidden := BitTypeEmpty
	revealed := RevealedBitBenign
	rarity := Junk
	symbol := " "
	if rand.Float32() < level.BitStreamChance {
		hidden = hiddenBits[rand.Intn(len(hiddenBits))]
		symbol = hidden.String()
		rarity = getRarity(level)
		next := rand.Float32()
		if next < level.BadBitChance {
			revealed = RevealedBitHarmful
			symbol = harmfulBits[rand.Intn(len(harmfulBits))]
		} else if next > (1 - level.GoodBitChance) {
			revealed = RevealedBitHelpful
			symbol = helpfulBits[rand.Intn(len(helpfulBits))]
		}
	}
	return Bits{hidden, revealed, rarity, symbol}
}

type World struct {
	Level     LevelConfig
	Loot      map[Coordinate]Loot
	BitStream map[Coordinate]Bits
}

func newWorld(level LevelConfig) World {

	bitStream := make(map[Coordinate]Bits)
	loot := make(map[Coordinate]Loot)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			loot[c] = Loot{
				Type: LootTypeEmpty,
			}
			bitStream[c] = getBit(level)
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
				loot[c] = Loot{dataType, getRarity(level)}
				filled++
			}
		}
	}

	return World{level, loot, bitStream}
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
			if c.Y < w.Level.Height {
				newStream[c] = val
			}
		}
		// Add a new row on top
		for x := 0; x < w.Level.Width; x++ {
			c := Coordinate{
				X: x,
				Y: 0,
			}
			newStream[c] = getBit(w.Level)
		}
	}
	w.BitStream = newStream
}

func (w *World) DidCollideWith(c Coordinate, bitType RevealedBitType) (float32, bool) {
	if b, ok := w.BitStream[c]; ok {
		return b.Threat(w.Level), b.Revealed == bitType
	}
	return 0, false
}

func (w *World) NeutralizeBit(c Coordinate) {
	if b, ok := w.BitStream[c]; ok {
		b.Revealed = RevealedBitBenign
		w.BitStream[c] = b
	}
}

func (w *World) DidCollideWithLoot(c Coordinate) bool {
	if l, ok := w.Loot[c]; ok {
		return l.Type != LootTypeEmpty
	}
	return false
}

func (w *World) ExtractLoot(c Coordinate) Loot {
	if l, ok := w.Loot[c]; ok {
		w.Loot[c] = Loot{Type: LootTypeEmpty}
		return l
	}
	return Loot{Type: LootTypeEmpty}
}
