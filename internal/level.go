package zen_doctor

import "fmt"

type LevelConfig struct {
	Level              Level
	Width              int
	Height             int
	ThreatDecay        float32
	MovementThreat     float32
	MaxThreat          float32
	FPS                int
	ViewDistX          int
	ViewDistY          int
	LootSpeed          float32
	LootDecay          float32
	BitStreamChance    float32
	GoodBitChance      float32
	BadBitChance       float32
	LootChanceByRarity map[Rarity]float32
	ThreatByRarity     map[Rarity]float32
	DataByRarity       map[Rarity]float32
	DataMultipliers    map[LootType]float32
	DataRequired       map[LootType]int
}

func (l LevelConfig) Name() string {
	return l.Level.String()
}

type Level int

const (
	Tutorial Level = iota
	Level1
	Level2
)

func (l Level) Equals(i int) bool {
	return int(l) == i
}

func (l Level) Inc() Level {
	return l + 1
}

func (l Level) Dec() Level {
	return l - 1
}

func (l Level) IsValid() bool {
	return l >= Tutorial && l <= Level2
}

func (l Level) String() string {
	switch l {
	case Tutorial:
		return "Level 0: Tutorial"
	case Level1:
		return "Level 1"
	case Level2:
		return "Level 2"
	default:
		return fmt.Sprintf("%d", int(l))
	}
}

func Levels() []Level {
	return []Level{
		Tutorial,
		Level1,
		Level2,
	}
}

func GetLevel(level Level) LevelConfig {
	switch level {
	case Tutorial:
		return LevelConfig{
			Level:           Tutorial,
			Width:           100,
			Height:          20,
			ThreatDecay:     -0.03,
			MovementThreat:  0.3,
			MaxThreat:       50,
			FPS:             2,
			ViewDistX:       6,
			ViewDistY:       3,
			LootSpeed:       1,
			LootDecay:       -0.3,
			BitStreamChance: 0.2,
			GoodBitChance:   0.02,
			BadBitChance:    0.1,
			LootChanceByRarity: map[Rarity]float32{
				Legendary: 0.005,
				Epic:      0.05,
				Rare:      0.3,
				Uncommon:  0.6,
				Common:    0.8,
				Junk:      1.0,
			},
			ThreatByRarity: map[Rarity]float32{
				Legendary: 15,
				Epic:      10,
				Rare:      6,
				Uncommon:  4,
				Common:    3,
				Junk:      2,
			},
			DataByRarity: map[Rarity]float32{
				Legendary: 1000,
				Epic:      100,
				Rare:      70,
				Uncommon:  40,
				Common:    25,
				Junk:      1,
			},
			DataMultipliers: map[LootType]float32{
				LootTypeDelta: 5,
			},
			DataRequired: map[LootType]int{
				LootTypeDelta: 21,
			},
		}
	default:
		return LevelConfig{}
	}
}

func (l LevelConfig) IsValid() bool {
	return l.Level.IsValid() && l.Width > 0 && l.Height > 0
}
