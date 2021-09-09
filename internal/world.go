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

type LootType int

const (
	LootTypeEmpty LootType = iota
	LootTypeDelta
	LootTypeOmega
	LootTypeSigma
	LootTypeLambda
)

func (lt LootType) SymbolForMode(mode CompatibilityMode) string {
	switch lt {
	case LootTypeDelta:
		return DeltaSymbol.ForMode(mode)
	case LootTypeOmega:
		return OmegaSymbol.ForMode(mode)
	case LootTypeSigma:
		return SigmaSymbol.ForMode(mode)
	case LootTypeLambda:
		return LambdaSymbol.ForMode(mode)
	default:
		return " "
	}
}

func getLootType(level *LevelConfig) LootType {
	picked := rand.Float32()
	lower := float32(0.0)
	for lt, chance := range level.LootTable {
		upper := lower + chance
		if picked >= lower && picked < upper {
			return lt
		}
		lower = upper
	}
	return level.DefaultLootType
}

type Loot struct {
	Type      LootType
	Rarity    Rarity
	Data      float32
	Integrity float32 // set to 1 initially, when it hits 0, the loot becomes worthless.
}

func newLoot(level *LevelConfig) Loot {
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

func (l *Loot) SymbolForMode(mode CompatibilityMode) (Color, string) {
	return l.Rarity.Color(), l.Type.SymbolForMode(mode)
}

func (l *Loot) tick(rate float32) {
	l.Integrity += rate
	if l.Integrity < 0 {
		l.Data = 0
		l.Rarity = Junk
		l.Type = LootTypeEmpty
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
	Level             *LevelConfig
	Loot              map[Coordinate]Loot
	Footprints        map[Coordinate]Footprint
	LootSpawnProgress float32
	Exit              *Coordinate
}

func newWorld(level *LevelConfig) World {
	world := World{
		Level:      level,
		Loot:       make(map[Coordinate]Loot),
		Footprints: make(map[Coordinate]Footprint),
	}
	world.spawnLoot(level.InitialLoot)
	return world
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
