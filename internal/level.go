package zen_doctor

import "fmt"

type LevelInfo struct {
	Level Level
	Width  int
	Height int
}

func (l LevelInfo) Name() string {
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
	return l+1
}

func (l Level) Dec() Level {
	return l-1
}

func (l Level) IsValid() bool {
	return l >= Tutorial && l <= Level2
}

func (l Level) String() string {
	switch l {
	case Tutorial:
		return "Tutorial"
	case Level1:
		return "Level 1"
	case Level2:
		return "Level 2"
	default:
		return fmt.Sprintf("%d", int(l))
	}
}

func Levels() []Level {
	return []Level {
		Tutorial,
		Level1,
		Level2,
	}
}

func GetLevel(level Level) LevelInfo {
	switch level {
	case Tutorial:
		return LevelInfo{
			Level: Tutorial,
			Width:  30,
			Height: 10,
		}
	default:
		return LevelInfo{}
	}
}

func (l LevelInfo) IsValid() bool {
	return l.Level.IsValid() && l.Width > 0 && l.Height > 0
}
