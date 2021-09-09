package zen_doctor

import (
	"math/rand"
	"time"
)

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
	Hidden         HiddenBitType
	Revealed       RevealedBitType
	Value          Rarity
	RevealedSymbol AnimatedSymbol
}

func (b Bits) ViewHidden() string {
	return b.Hidden.String()
}
func (b Bits) ViewRevealed(mode CompatibilityMode) (Color, string) {
	color := LightGray
	switch b.Revealed {
	case RevealedBitHelpful:
		color = Green
	case RevealedBitHarmful:
		color = Red
	}
	if b.RevealedSymbol != nil {
		return color, b.RevealedSymbol.ForMode(mode)
	}
	return color, b.Hidden.String()
}

// Threat returns the magnitude of the threat based on the value.
// you still need to multiply by -1 if it's helpful
func (b Bits) Threat(level *LevelConfig) float32 {
	if threat, ok := level.ThreatByRarity[b.Value]; ok {
		return threat
	}
	return 0
}

var hiddenBits = []HiddenBitType{
	BitTypeZero,
	BitTypeOne,
}

func getBit(level *LevelConfig) Bits {
	if rand.Float32() < level.BitStreamChance {
		hidden := hiddenBits[rand.Intn(len(hiddenBits))]
		rarity := getRarity(level)
		next := rand.Float32()
		revealed := RevealedBitBenign
		if next < level.BadBitChance {
			revealed = RevealedBitHarmful
		} else if next > (1 - level.GoodBitChance) {
			revealed = RevealedBitHelpful
		}
		revealedSymbol := GenerateNoiseSymbolFor(revealed, rarity)
		return Bits{hidden, revealed, rarity, revealedSymbol}
	}
	return Bits{BitTypeEmpty, RevealedBitBenign, Junk, nil}
}

type BitStream struct {
	level  *LevelConfig
	stream map[Coordinate]Bits
}

func newBitStream(level *LevelConfig) BitStream {
	stream := make(map[Coordinate]Bits)
	for x := 0; x < level.Width; x++ {
		for y := 0; y < level.Height; y++ {
			c := Coordinate{x, y}
			stream[c] = getBit(level)
		}
	}
	return BitStream{
		level:  level,
		stream: stream,
	}
}

func (bs *BitStream) TickAnimations() {
	for c := range bs.stream {
		if bs.stream[c].RevealedSymbol != nil {
			bs.stream[c].RevealedSymbol.Tick()
		}
	}
}

func (bs *BitStream) DidCollideWithBit(level *LevelConfig, c Coordinate, bitType RevealedBitType) (float32, bool) {
	if b, ok := bs.stream[c]; ok {
		return b.Threat(level), b.Revealed == bitType
	}
	return 0, false
}

func (bs *BitStream) NeutralizeBit(c Coordinate) {
	if b, ok := bs.stream[c]; ok {
		b.Revealed = RevealedBitBenign
		bs.stream[c] = b
	}
}

func shiftBitStream(dir Direction, stream *BitStream) {
	newStream := make(map[Coordinate]Bits)

	// Move all existing bits
	for coord, val := range stream.stream {
		var c Coordinate
		switch dir {
		case MoveDown:
			c = Coordinate{X: coord.X, Y: coord.Y + 1}
		case MoveUp:
			c = Coordinate{X: coord.X, Y: coord.Y - 1}
		case MoveLeft:
			c = Coordinate{X: coord.X - 1, Y: coord.Y}
		case MoveRight:
			c = Coordinate{X: coord.X + 1, Y: coord.Y}
		case MoveUpRight:
			c = Coordinate{X: coord.X + 1, Y: coord.Y - 1}
		case MoveUpLeft:
			c = Coordinate{X: coord.X - 1, Y: coord.Y - 1}
		case MoveDownRight:
			c = Coordinate{X: coord.X + 1, Y: coord.Y + 1}
		case MoveDownLeft:
			c = Coordinate{X: coord.X - 1, Y: coord.Y + 1}
		}

		if c.X < stream.level.Width && c.X >= 0 && c.Y < stream.level.Height && c.Y >= 0 {
			newStream[c] = val
		} else {
			// replace this bit
			if c.X >= stream.level.Width {
				c.X = 0
			} else if c.X < 0 {
				c.X = stream.level.Width - 1
			}
			if c.Y >= stream.level.Height {
				c.Y = 0
			} else if c.Y < 0 {
				c.Y = stream.level.Height - 1
			}
			newStream[c] = getBit(stream.level)
		}
	}
	stream.stream = newStream
}

type BitStreamUpdater interface {
	Tick(stream *BitStream)
}

type linearBitStream struct {
	dir Direction
}

func newLinearBitStream(dir Direction) BitStreamUpdater {
	return &linearBitStream{dir}
}

func (b *linearBitStream) Tick(stream *BitStream) {
	shiftBitStream(b.dir, stream)
}

type shiftingBitStream struct {
	steps      []bitStreamStep
	current    int
	lastUpdate time.Time
}

type bitStreamStep struct {
	dir   Direction
	delay time.Duration
}

func newShiftingBitStream(steps ...bitStreamStep) BitStreamUpdater {
	return &shiftingBitStream{
		steps:      steps,
		lastUpdate: time.Now(),
	}
}

func (s *shiftingBitStream) Tick(stream *BitStream) {
	now := time.Now()
	if now.Sub(s.lastUpdate) > s.steps[s.current].delay {
		s.current++
		if s.current >= len(s.steps) {
			s.current = 0
		}
		s.lastUpdate = now
	}
	shiftBitStream(s.steps[s.current].dir, stream)
}

func rotatingBitStream(vertical, horizontal, diagonal time.Duration) []bitStreamStep {
	return []bitStreamStep{
		{
			dir:   MoveDown,
			delay: vertical,
		}, {
			dir:   MoveDownLeft,
			delay: diagonal,
		}, {
			dir:   MoveLeft,
			delay: horizontal,
		}, {
			dir:   MoveUpLeft,
			delay: diagonal,
		}, {
			dir:   MoveUp,
			delay: vertical,
		}, {
			dir:   MoveUpRight,
			delay: diagonal,
		}, {
			dir:   MoveRight,
			delay: horizontal,
		}, {
			dir:   MoveDownRight,
			delay: diagonal,
		},
	}
}

func zigZagBitStream(cardinal, diagonal time.Duration) []bitStreamStep {
	return []bitStreamStep{
		{
			dir:   MoveDown,
			delay: cardinal,
		}, {
			dir:   MoveDownLeft,
			delay: diagonal,
		}, {
			dir:   MoveDown,
			delay: cardinal,
		}, {
			dir:   MoveDownRight,
			delay: diagonal,
		},
	}
}
