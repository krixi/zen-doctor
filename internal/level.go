package zen_doctor

import "fmt"

type LevelSettings struct {
	Level          Level
	Width          int
	Height         int
	ThreatDecay    float32
	MovementThreat float32
	MaxThreat      float32
	FPS            int
	ViewDistX      int
	ViewDistY      int
	LootSpeed      float32
	LootDecay      float32
	DataRequired   map[LootType]int
}

func (l LevelSettings) Name() string {
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

func GetLevel(level Level) LevelSettings {
	switch level {
	case Tutorial:
		return LevelSettings{
			Level:          Tutorial,
			Width:          100,
			Height:         20,
			ThreatDecay:    -0.03,
			MovementThreat: 0.3,
			MaxThreat:      50,
			FPS:            2,
			ViewDistX:      6,
			ViewDistY:      3,
			LootSpeed:      1,
			LootDecay:      -0.3,
			DataRequired: map[LootType]int{
				LootTypeDelta: 1,
			},
		}
	default:
		return LevelSettings{}
	}
}

func (l LevelSettings) IsValid() bool {
	return l.Level.IsValid() && l.Width > 0 && l.Height > 0
}
