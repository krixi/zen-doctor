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

func (lt LootType) WithRarity(rarity Rarity) string {
	msg := lt.String()
	switch rarity {
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

func getLootType(level LevelConfig) LootType {

	checker := func(lootType LootType) bool {
		if chance, ok := level.LootTable[lootType]; ok {
			return rand.Float32() < chance
		}
		return false
	}

	for _, want := range level.WinConditions {
		if checker(want.Type) {
			return want.Type
		}
	}
	for _, want := range level.Bonus {
		if checker(want) {
			return want
		}
	}
	return level.DefaultLootType
}

type Loot struct {
	Type      LootType
	Rarity    Rarity
	Data      float32
	Integrity float32 // set to 1 initially, when it hits 0, the loot becomes worthless.
}

func newLoot(level LevelConfig) Loot {
	lootType := getLootType(level)
	rarity := getRarity(level)
	data := level.DataByRarity[rarity] * level.DataMultipliers[lootType]
	return Loot{
		Type:      lootType,
		Rarity:    rarity,
		Data:      data,
		Integrity: 1,
	}
}

func (l *Loot) String() string {
	return l.Type.WithRarity(l.Rarity)
}

func (l *Loot) tick(rate float32) {
	l.Integrity += rate
	if l.Integrity < 0 {
		l.Data = 0
		l.Rarity = Junk
		l.Type = LootTypeEmpty
	}
}

func (l *Loot) WithIntegrity(msg string) string {
	if l.Integrity > 0.33 {
		return WithColor(White, msg)
	}
	if l.Integrity > 0.15 {
		return WithColor(Yellow, msg)
	}
	return WithColor(Red, msg)
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

func (rbt RevealedBitType) SymbolByRarity(rarity Rarity, defaultStr string) string {
	switch rbt {
	case RevealedBitHelpful:
		switch rarity {
		case Legendary:
			return RuneWSymbol
		case Epic:
			return RuneFSymbol
		case Rare:
			return RuneIngSymbol
		case Uncommon:
			return RuneSowiloSymbol
		case Common:
			return RuneLongOSymbol
		case Junk:
			return RuneShortOSymbol
		}
	case RevealedBitHarmful:
		switch rarity {
		case Legendary:
			return RuneCelacSymbol
		case Epic:
			return RuneTvimadurSymbol
		case Rare:
			return RuneEolhxSymbol
		case Uncommon:
			return PsiSymbol
		case Common:
			return DaggerSymbol
		case Junk:
			return KoppaSymbol
		}
	}
	return defaultStr
}

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

func getBit(level LevelConfig) Bits {
	hidden := BitTypeEmpty
	revealed := RevealedBitBenign
	rarity := Junk
	symbol := " "
	if rand.Float32() < level.BitStreamChance {
		hidden = hiddenBits[rand.Intn(len(hiddenBits))]
		rarity = getRarity(level)
		next := rand.Float32()
		if next < level.BadBitChance {
			revealed = RevealedBitHarmful
		} else if next > (1 - level.GoodBitChance) {
			revealed = RevealedBitHelpful
		}
		symbol = revealed.SymbolByRarity(rarity, hidden.String())
	}
	return Bits{hidden, revealed, rarity, symbol}
}

type World struct {
	Level             LevelConfig
	Loot              map[Coordinate]Loot
	BitStream         map[Coordinate]Bits
	LootSpawnProgress float32
	Exit              *Coordinate
}

func newWorld(level LevelConfig) World {

	bitStream := make(map[Coordinate]Bits)
	loot := make(map[Coordinate]Loot)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			bitStream[c] = getBit(level)
		}
	}

	world := World{
		Level:     level,
		Loot:      loot,
		BitStream: bitStream,
	}
	world.spawnLoot(level.InitialLoot)
	return world
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

func (w *World) DidCollideWithBit(c Coordinate, bitType RevealedBitType) (float32, bool) {
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

func (w *World) spawnLoot(n int) {
	filled := 0
	for filled < n {
		// make sure it's empty first
		x := rand.Intn(w.Level.Width - 1)
		y := rand.Intn(w.Level.Height - 1)
		c := Coordinate{x, y}

		// even though it's a sparse map, this should work due to default types in go :squint:
		if w.Loot[c].Type == LootTypeEmpty {
			w.Loot[c] = newLoot(w.Level)
			filled++
		}
	}
}

func (w *World) TickLootAt(c Coordinate, rate float32) {
	if loot, ok := w.Loot[c]; ok {
		loot.tick(rate)
		w.Loot[c] = loot
	}
}

func (w *World) TickLoot() {
	// tick all existing loot
	newLoot := make(map[Coordinate]Loot)
	for c, loot := range w.Loot {
		loot.tick(w.Level.LootDecayRate)
		if loot.Type != LootTypeEmpty {
			newLoot[c] = loot
		}
	}
	w.Loot = newLoot

	// make sure there's always at least one loot in the world
	if len(newLoot) == 0 {
		w.spawnLoot(1)
	}

	// spawn new loot if needed
	w.LootSpawnProgress += w.Level.LootSpawnRate
	if w.LootSpawnProgress > 1 {
		w.LootSpawnProgress = 0
		w.spawnLoot(1)
	}
}

func (w *World) UnlockExit() {
	if w.Exit == nil {
		x := rand.Intn(w.Level.Width - 1)
		y := rand.Intn(w.Level.Height - 1)
		w.Exit = &Coordinate{x, y}
	}
}

func (w *World) DidCollideWithExit(c Coordinate) bool {
	if w.Exit == nil {
		return false
	}
	return w.Exit.X == c.X && w.Exit.Y == c.Y
}
