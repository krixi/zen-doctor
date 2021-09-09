package zen_doctor

import "math/rand"

type CompatibilityMode int

const (
	CompatibilityAny CompatibilityMode = iota
	CompatibilityLatin
	CompatibilityAscii
)

type AnimatedSymbol interface {
	Symbol
	Tick()
}

type NoisySymbol struct {
	Base     symbol
	Noise    []symbol
	Chance   float32
	selected int
	useNoise bool
}

func (s *NoisySymbol) Tick() {
	if rand.Float32() < s.Chance {
		// select noise
		s.useNoise = true
		s.selected = rand.Intn(len(s.Noise))
	} else {
		s.useNoise = false
	}
}

func (s *NoisySymbol) ForMode(mode CompatibilityMode) string {
	if s.useNoise {
		return s.Noise[s.selected].ForMode(mode)
	} else {
		return s.Base.ForMode(mode)
	}
}

type LoopingSymbol struct {
	Frames  []symbol
	Current int
}

func (s *LoopingSymbol) Tick() {
	s.Current++
	if s.Current >= len(s.Frames) {
		s.Current = 0
	}
}

func (s *LoopingSymbol) ForMode(mode CompatibilityMode) string {
	return s.Frames[s.Current].ForMode(mode)
}

var AnimatedExit = LoopingSymbol{
	Frames: []symbol{
		//{
		//	Runic: `ᛟ`,
		//	Latin: `Ξ`,
		//	ASCII: `.`,
		//},
		{
			Runic: `▁`,
			Latin: `▁`,
			ASCII: `o`,
		},
		{
			Runic: `▏`,
			Latin: `▏`,
			ASCII: `O`,
		},
		{
			Runic: `▔`,
			Latin: `▔`,
			ASCII: `o`,
		},
		{
			Runic: `▕`,
			Latin: `▕`,
			ASCII: `.`,
		},
	},
}

type Symbol interface {
	ForMode(CompatibilityMode) string
}

type symbol struct {
	Runic string
	Latin string
	ASCII string
}

func (s *symbol) ForMode(mode CompatibilityMode) string {
	switch mode {
	case CompatibilityAscii:
		return s.ASCII
	case CompatibilityLatin:
		return s.Latin
	default:
		return s.Runic
	}
}

var PlayerSymbol = symbol{
	Runic: `Ȣ`,
	Latin: `Ȣ`,
	ASCII: `@`,
}

// Helpful stream
var GoodBit1 = symbol{
	Runic: `ᚭ`,
	Latin: `1`,
	ASCII: `1`,
}
var GoodBit2 = symbol{
	Runic: `ᚬ`,
	Latin: `2`,
	ASCII: `2`,
}
var GoodBit3 = symbol{
	Runic: `ᛊ`,
	Latin: `3`,
	ASCII: `3`,
}
var GoodBit4 = symbol{
	Runic: `ᛝ`,
	Latin: `4`,
	ASCII: `4`,
}
var GoodBit5 = symbol{
	Runic: `ᚠ`,
	Latin: `5`,
	ASCII: `5`,
}
var GoodBit6 = symbol{
	Runic: `ᚥ`,
	Latin: `6`,
	ASCII: `6`,
}

// Harmful stream
var BadBit1 = symbol{
	Runic: `ϟ`,
	Latin: `a`,
	ASCII: `a`,
}
var BadBit2 = symbol{
	Runic: `†`,
	Latin: `b`,
	ASCII: `b`,
}
var BadBit3 = symbol{
	Runic: `ψ`,
	Latin: `c`,
	ASCII: `c`,
}
var BadBit4 = symbol{
	Runic: `ᛉ`,
	Latin: `d`,
	ASCII: `d`,
}
var BadBit5 = symbol{
	Runic: `ᛯ`,
	Latin: `e`,
	ASCII: `e`,
}
var BadBit6 = symbol{
	Runic: `ᛤ`,
	Latin: `f`,
	ASCII: `f`,
}

var noise = []symbol{
	{
		Runic: `▖`,
		Latin: `▖`,
		ASCII: `#`,
	},
	{
		Runic: `▗`,
		Latin: `▗`,
		ASCII: `$`,
	},
	{
		Runic: `▘`,
		Latin: `▘`,
		ASCII: `!`,
	},
	{
		Runic: `▙`,
		Latin: `▙`,
		ASCII: `&`,
	},
	{
		Runic: `▚`,
		Latin: `▚`,
		ASCII: `-`,
	},
	{
		Runic: `▛`,
		Latin: `▛`,
		ASCII: `/`,
	},
	{
		Runic: `▜`,
		Latin: `▜`,
		ASCII: `\`,
	},
	{
		Runic: `▝`,
		Latin: `▝`,
		ASCII: `{`,
	},
	{
		Runic: `▞`,
		Latin: `▞`,
		ASCII: `}`,
	},
	{
		Runic: `▟`,
		Latin: `▟`,
		ASCII: `~`,
	},
}

var badBitSymbolsByRarity = map[Rarity]symbol{
	Legendary: BadBit6,
	Epic:      BadBit5,
	Rare:      BadBit4,
	Uncommon:  BadBit3,
	Common:    BadBit2,
	Junk:      BadBit1,
}

var goodBitSymbolsByRarity = map[Rarity]symbol{
	Legendary: GoodBit6,
	Epic:      GoodBit5,
	Rare:      GoodBit4,
	Uncommon:  GoodBit3,
	Common:    GoodBit2,
	Junk:      GoodBit1,
}

var noiseChanceByRarity = map[Rarity]float32{
	Legendary: 0.20,
	Epic:      0.15,
	Rare:      0.12,
	Uncommon:  0.08,
	Common:    0.05,
	Junk:      0.03,
}

func GenerateNoiseSymbolFor(bitType RevealedBitType, rarity Rarity) AnimatedSymbol {
	noiseSymbol := NoisySymbol{
		Noise:  noise,
		Chance: noiseChanceByRarity[rarity],
	}
	switch bitType {
	case RevealedBitHarmful:
		noiseSymbol.Base = badBitSymbolsByRarity[rarity]
	case RevealedBitHelpful:
		noiseSymbol.Base = goodBitSymbolsByRarity[rarity]
	default:
		return nil
	}
	return &noiseSymbol
}

// Loot symbols
var DeltaSymbol = symbol{
	Runic: `Δ`,
	Latin: `Δ`,
	ASCII: `W`,
}
var LambdaSymbol = symbol{
	Runic: `λ`,
	Latin: `λ`,
	ASCII: `X`,
}
var SigmaSymbol = symbol{
	Runic: `Σ`,
	Latin: `Σ`,
	ASCII: `Y`,
}
var OmegaSymbol = symbol{
	Runic: `Ω`,
	Latin: `Ω`,
	ASCII: `Z`,
}

var ProgressBarSymbol = symbol{
	Runic: `█`,
	Latin: `█`,
	ASCII: `#`,
}

const (
	QuestionSymbol  = `?`
	FootprintSymbol = `.`
)
