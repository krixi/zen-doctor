package zen_doctor

type CompatibilityMode int

const (
	CompatibilityAny CompatibilityMode = iota
	CompatibilityLatin
	CompatibilityAscii
)

type Symbol struct {
	Runic string
	Latin string
	ASCII string
}

func (s Symbol) ForMode(mode CompatibilityMode) string {
	switch mode {
	case CompatibilityAscii:
		return s.ASCII
	case CompatibilityLatin:
		return s.Latin
	default:
		return s.Runic
	}
}

var PlayerSymbolS = Symbol{
	Runic: `Ȣ`,
	Latin: `Ȣ`,
	ASCII: `@`,
}

// Helpful bits
var GoodBit1 = Symbol{
	Runic: `ᚭ`,
	Latin: `1`,
	ASCII: `1`,
}
var GoodBit2 = Symbol{
	Runic: `ᚬ`,
	Latin: `2`,
	ASCII: `2`,
}
var GoodBit3 = Symbol{
	Runic: `ᛊ`,
	Latin: `3`,
	ASCII: `3`,
}
var GoodBit4 = Symbol{
	Runic: `ᛝ`,
	Latin: `4`,
	ASCII: `4`,
}
var GoodBit5 = Symbol{
	Runic: `ᚠ`,
	Latin: `5`,
	ASCII: `5`,
}
var GoodBit6 = Symbol{
	Runic: `ᚥ`,
	Latin: `6`,
	ASCII: `6`,
}

// Harmful bits
var BadBit1 = Symbol{
	Runic: `ϟ`,
	Latin: `a`,
	ASCII: `a`,
}
var BadBit2 = Symbol{
	Runic: `†`,
	Latin: `b`,
	ASCII: `b`,
}
var BadBit3 = Symbol{
	Runic: `ψ`,
	Latin: `c`,
	ASCII: `c`,
}
var BadBit4 = Symbol{
	Runic: `ᛉ`,
	Latin: `d`,
	ASCII: `d`,
}
var BadBit5 = Symbol{
	Runic: `ᛯ`,
	Latin: `e`,
	ASCII: `e`,
}
var BadBit6 = Symbol{
	Runic: `ᛤ`,
	Latin: `f`,
	ASCII: `f`,
}

// Loot symbols
var DeltaSymbol = Symbol{
	Runic: `Δ`,
	Latin: `Δ`,
	ASCII: `W`,
}
var LambdaSymbol = Symbol{
	Runic: `λ`,
	Latin: `λ`,
	ASCII: `X`,
}
var SigmaSymbol = Symbol{
	Runic: `Σ`,
	Latin: `Σ`,
	ASCII: `Y`,
}
var OmegaSymbol = Symbol{
	Runic: `Ω`,
	Latin: `Ω`,
	ASCII: `Z`,
}

var ExitSymbol = Symbol{
	Runic: `ᛟ`,
	Latin: `*`,
	ASCII: `*`,
}

var ProgressBarSymbol = Symbol{
	Runic: `█`,
	Latin: `█`,
	ASCII: `#`,
}

const (
	QuestionSymbol  = `?`
	FootprintSymbol = `.`
)
