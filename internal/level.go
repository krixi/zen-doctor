package zen_doctor

type LevelInfo struct {
	Name   string
	Width  int
	Height int
}

type Level int

const (
	Tutorial Level = iota
	Level1
	Level2
	Level3
	Level4
	Level5
)

func GetLevel(level Level) LevelInfo {
	switch level {
	case Tutorial:
		return LevelInfo{
			Name:   "Tutorial",
			Width:  30,
			Height: 10,
		}
	default:
		return LevelInfo{}
	}
}

func (l LevelInfo) IsValid() bool {
	return l.Name != "" && l.Width > 0 && l.Height > 0
}
