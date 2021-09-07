package zen_doctor

import "fmt"

type LevelConfig struct {
	Level              Level
	Width              int // Size of map
	Height             int
	ViewDistX          int // How far the player can see
	ViewDistY          int
	FPS                float32 // How fast the bit stream renders
	ThreatDecay        float32 // how fast threat meter decays
	MovementThreat     float32 // how much threat per step
	MaxThreat          float32
	BitStreamChance    float32              // % chance a bit will appear
	GoodBitChance      float32              // % chance a bit will be good
	BadBitChance       float32              // % chance a bit will be good
	ThreatByRarity     map[Rarity]float32   // amount of threat generated by bits, based on their rarity
	LeaveSpeed         float32              // how fast the player can leave the room
	LeaveSpeedDecay    float32              // how fast the player can leave the room
	InitialLoot        int                  // how much loot is spawned into the world when it's loaded
	DefaultLootType    LootType             // in case we need to spawn loot and the loot table comes up empty
	LootSpeed          float32              // how fast loot meter fills when looting
	LootSpeedDecay     float32              // how fast loot meter falls when looting is interrupted
	LootSpawnRate      float32              // how fast new loot is spawned into the world
	LootDecayRate      float32              // how fast loot in the world decays
	LootTable          map[LootType]float32 // % chance for loot to be spawned
	LootChanceByRarity map[Rarity]float32   // % chance for loot to be a specific rarity
	DataByRarity       map[Rarity]float32   // how much data is worth, by loot rarity
	DataMultipliers    map[LootType]float32 // multiplier for how much data is worth, by loot type
	WinConditions      []WinCondition       // what is required to unlock the exit to this room
	Bonus              []LootType           // the loot types considered bonus for this level
}

type WinCondition struct {
	Type   LootType
	Amount float32
}

func (w WinCondition) IsMet(state *GameState) bool {
	if amount, ok := state.player.DataCollected[w.Type]; ok {
		return amount >= w.Amount
	}
	return false
}

func (l LevelConfig) Name() string {
	return l.Level.String()
}

type Level int

const (
	Tutorial Level = iota
	Level1
	//Level2
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
	return l >= Tutorial && l <= Level1
}

func (l Level) String() string {
	switch l {
	case Tutorial:
		return "Level 0: Tutorial"
	case Level1:
		return "Level 1"
	//case Level2:
	//	return "Level 2"
	default:
		return fmt.Sprintf("%d", int(l))
	}
}

func defaultLevel() LevelConfig {
	return LevelConfig{
		Level:           Tutorial,
		Width:           100,
		Height:          20,
		ViewDistX:       6,
		ViewDistY:       3,
		FPS:             1.25,
		ThreatDecay:     -0.03,
		MovementThreat:  0.3,
		MaxThreat:       50,
		BitStreamChance: 0.2,
		GoodBitChance:   0.02,
		BadBitChance:    0.1,
		ThreatByRarity: map[Rarity]float32{
			Legendary: 15,
			Epic:      10,
			Rare:      6,
			Uncommon:  4,
			Common:    3,
			Junk:      2,
		},
		LeaveSpeed:      2,
		LeaveSpeedDecay: -0.1,
		InitialLoot:     1,
		DefaultLootType: LootTypeDelta,
		LootSpeed:       1,
		LootSpeedDecay:  -0.3,
		LootSpawnRate:   0.003,
		LootDecayRate:   -0.001,
		LootTable: map[LootType]float32{
			LootTypeDelta:  0.25,
			LootTypeLambda: 0.10,
			LootTypeSigma:  0.10,
			LootTypeOmega:  0.10,
		},
		LootChanceByRarity: map[Rarity]float32{
			Legendary: 0.005,
			Epic:      0.05,
			Rare:      0.3,
			Uncommon:  0.6,
			Common:    0.8,
			Junk:      1.0,
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
			LootTypeDelta:  1,
			LootTypeLambda: 1,
			LootTypeSigma:  1,
			LootTypeOmega:  1,
		},
		WinConditions: []WinCondition{
			{
				Type:   LootTypeDelta,
				Amount: 100,
			},
		},
		Bonus: []LootType{LootTypeLambda, LootTypeSigma, LootTypeOmega},
	}
}

func GetLevel(level Level) LevelConfig {
	l := defaultLevel()
	l.Level = level

	// customize defaults for the requested level
	switch level {
	case Level1:
		l.FPS = 2
		l.WinConditions = []WinCondition{
			{
				Type:   LootTypeDelta,
				Amount: 200,
			},
			{
				Type:   LootTypeLambda,
				Amount: 50,
			},
		}
		l.Bonus = []LootType{LootTypeSigma, LootTypeOmega}
		l.LootTable[LootTypeLambda] = 0.25
		l.DataMultipliers[LootTypeDelta] = 1.2
	}
	return l
}

func (l LevelConfig) IsValid() bool {
	return l.Level.IsValid() && l.Width > 0 && l.Height > 0
}
