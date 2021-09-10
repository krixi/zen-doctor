package zen_doctor

import (
	"math"
	"math/rand"
)

type Coordinate struct {
	X int
	Y int
}

func (c Coordinate) InRange(radius float64, other Coordinate) bool {
	rx := float64(c.X - other.X)
	ry := float64(c.Y-other.Y) * 2 // to compensate for terminal character sizes
	return math.Sqrt(rx*rx+ry*ry) < radius
}

func (c Coordinate) Equals(other Coordinate) bool {
	return c.X == other.X && c.Y == other.Y
}

type Direction int

const (
	MoveUp Direction = iota // cardinal
	MoveDown
	MoveLeft
	MoveRight
	MoveUpLeft // diagonal
	MoveUpRight
	MoveDownLeft
	MoveDownRight
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

func (r Rarity) Color() Color {
	switch r {
	case Junk:
		return LightGray
	case Common:
		return White
	case Uncommon:
		return Green
	case Rare:
		return Blue
	case Epic:
		return Purple
	case Legendary:
		return Orange
	}
	return DarkGray
}

func getRarity(level *LevelConfig) Rarity {
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

type DataKind int

const (
	DataKindNone DataKind = iota
	DataKindDelta
	DataKindLambda
	DataKindSigma
	DataKindOmega
)

func (k DataKind) ForMode(mode CompatibilityMode) string {
	switch k {
	case DataKindDelta:
		return DeltaSymbol.ForMode(mode)
	case DataKindOmega:
		return OmegaSymbol.ForMode(mode)
	case DataKindSigma:
		return SigmaSymbol.ForMode(mode)
	case DataKindLambda:
		return LambdaSymbol.ForMode(mode)
	default:
		return ` `
	}
}

type PowerUpKind int

const (
	PowerUpNone PowerUpKind = iota
	PowerUpVisionRange
	PowerUpThreatDecay
	PowerUpBadBitImmunity
	PowerUpBadBitsAreGood
	PowerUpLootSpeed
)

func (k PowerUpKind) ForMode(mode CompatibilityMode) string {
	switch k {
	case PowerUpVisionRange:
		return VisionRangeSymbol.ForMode(mode)
	case PowerUpThreatDecay:
		return ThreatDecaySymbol.ForMode(mode)
	case PowerUpBadBitImmunity:
		return BadBitImmunitySymbol.ForMode(mode)
	case PowerUpBadBitsAreGood:
		return BadBitsAreGoodSymbol.ForMode(mode)
	case PowerUpLootSpeed:
		return LootSpeedSymbol.ForMode(mode)
	default:
		return ` `
	}
}

type lootTable []lootOption

func (lt *lootTable) Chance(idx int) float32 {
	if idx >= len(*lt) {
		return 0
	}
	return (*lt)[idx].Chance
}

func (lt *lootTable) Len() int {
	return len(*lt)
}

type lootOption struct {
	Data    DataKind
	PowerUp PowerUpKind
	Chance  float32
}

type LootKind int

const (
	LootEmpty LootKind = iota
	LootData
	LootPowerUp
)

type Loot struct {
	Kind        LootKind
	Rarity      Rarity
	PowerUpKind PowerUpKind
	DataKind    DataKind
	Data        float32
	Integrity   float32 // set to 1 initially, when it hits 0, the loot becomes worthless and disappears
}

func newLoot(kind LootKind, level *LevelConfig) Loot {
	rarity := getRarity(level)
	loot := Loot{
		Kind:      kind,
		Rarity:    rarity,
		Integrity: 1,
	}
	switch kind {
	case LootData:
		dataKind := level.DataLootTable[pickOne(&level.DataLootTable)].Data
		data := level.DataByRarity[rarity] * level.DataMultipliers[dataKind]
		loot.Data, loot.DataKind = data, dataKind
	case LootPowerUp:
		powerUpKind := level.PowerUpLootTable[pickOne(&level.PowerUpLootTable)].PowerUp
		loot.PowerUpKind = powerUpKind
	}
	return loot
}

func (l *Loot) SymbolForMode(mode CompatibilityMode) (Color, string) {
	switch l.Kind {
	case LootData:
		return l.Rarity.Color(), l.DataKind.ForMode(mode)
	case LootPowerUp:
		return l.Rarity.Color(), l.PowerUpKind.ForMode(mode)
	}
	return l.Rarity.Color(), ` `
}

func (l *Loot) tick(rate float32) {
	l.Integrity += rate
	if l.Integrity < 0 {
		l.Data = 0
		l.Rarity = Junk
		l.Kind = LootEmpty
	}
}

func (l *Loot) WithIntegrity(msg string) (Color, string) {
	if l.Integrity > 0.33 {
		return White, msg
	}
	if l.Integrity > 0.15 {
		return Yellow, msg
	}
	return Red, msg
}

type Footprint struct {
	Intensity float32
}

func (f *Footprint) tick(rate float32) {
	f.Intensity += rate
}

func (f *Footprint) WithIntensity() (Color, string) {
	// linearly interpolate 100 (x0) -> 0 (x1) across 255 (y0) -> 235 (y1) :elahmm:
	y := (float32(White)*-f.Intensity + float32(DarkGray)*(f.Intensity-100)) / -100
	return Color(int(y)), FootprintSymbol
}

type World struct {
	Level                *LevelConfig
	Loot                 map[Coordinate]Loot
	Footprints           map[Coordinate]Footprint
	DataSpawnProgress    float32
	PowerUpSpawnProgress float32
	Exit                 *Coordinate
}

func newWorld(level *LevelConfig) World {
	world := World{
		Level:      level,
		Loot:       make(map[Coordinate]Loot),
		Footprints: make(map[Coordinate]Footprint),
	}
	world.spawnLoot(level.InitialData, LootData)
	world.spawnLoot(level.InitialPowerUps, LootPowerUp)
	return world
}

func (w *World) DidCollideWithLoot(c Coordinate) bool {
	if l, ok := w.Loot[c]; ok {
		return l.Kind != LootEmpty
	}
	return false
}

func (w *World) ExtractLoot(c Coordinate) Loot {
	if l, ok := w.Loot[c]; ok {
		w.Loot[c] = Loot{Kind: LootEmpty}
		return l
	}
	return Loot{
		Kind: LootEmpty,
	}
}

func (w *World) spawnLoot(n int, kind LootKind) {
	filled := 0
	for filled < n {
		// make sure it's empty first
		x := rand.Intn(w.Level.Width - 1)
		y := rand.Intn(w.Level.Height - 1)
		c := Coordinate{x, y}

		// even though it's a sparse map, this should work due to default types in go :squint:
		if w.Loot[c].Kind == LootEmpty {
			w.Loot[c] = newLoot(kind, w.Level)
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
		loot.tick(w.Level.DataDecayRate)
		if loot.Kind != LootEmpty {
			newLoot[c] = loot
		}
	}
	w.Loot = newLoot

	// make sure there's always at least one loot in the world
	if len(newLoot) == 0 {
		w.spawnLoot(1, LootData)
	}

	// spawn new loot if needed
	w.DataSpawnProgress += w.Level.DataSpawnRate
	if w.DataSpawnProgress > 1 {
		w.DataSpawnProgress = 0
		w.spawnLoot(1, LootData)
	}
	w.PowerUpSpawnProgress += w.Level.PowerUpSpawnRate
	if w.PowerUpSpawnProgress > 1 {
		w.PowerUpSpawnProgress = 0
		w.spawnLoot(1, LootPowerUp)
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

func (w *World) Visited(c Coordinate) {
	w.Footprints[c] = Footprint{Intensity: 100}
}

func (w *World) TickFootprints() {
	newFootprints := make(map[Coordinate]Footprint)
	for c, footprint := range w.Footprints {
		footprint.tick(w.Level.FootprintDecay)
		if footprint.Intensity > 0 {
			newFootprints[c] = footprint
		}
	}
	w.Footprints = newFootprints
}
