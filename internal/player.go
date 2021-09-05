package zen_doctor

const (
	PlayerSymbol        = "Ȣ"
	DeltaSymbol         = "Δ"
	OmegaSymbol         = "Ω"
	PhiSymbol           = "Φ"
	SigmaSymbol         = "Σ"
	LambdaSymbol        = "λ"
	PsiSymbol           = "ψ"
	KoppaSymbol         = "ϟ"
	SampiSymbol         = "Ϡ"
	ZheSymbol           = "Ж"
	ShchaSymbol         = "Щ"
	DaggerSymbol        = "†"
	ReferenceMarkSymbol = "※"
	ShrugSymbol         = "ツ"
)

var Symbols = []string{
	PlayerSymbol,
	DeltaSymbol,
	OmegaSymbol,
	PhiSymbol,
	SigmaSymbol,
	LambdaSymbol,
	PsiSymbol,
	KoppaSymbol,
	SampiSymbol,
	ZheSymbol,
	ShchaSymbol,
	DaggerSymbol,
	ReferenceMarkSymbol,
	ShrugSymbol,
}

type Player struct {
	Location  Coordinate
	Threat    int
	ViewDistX int
	ViewDistY int
}

func newPlayer(loc Coordinate) Player {
	return Player{
		Location:  loc,
		Threat:    0,
		ViewDistX: 4,
		ViewDistY: 2,
	}
}
