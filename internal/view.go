package zen_doctor

import (
	"fmt"
	"strings"
)

func CalculateViewPosition(width, height, screenWidth, screenHeight int) (int, int, int, int) {
	x1 := screenWidth/2 - width/2
	y1 := screenHeight/2 - height/2
	x2 := x1 + width
	y2 := y1 + height
	return x1, y1, x2, y2
}

type Color int

const (
	DarkBlue    Color = 20
	LightBlue   Color = 81
	Teal        Color = 85
	Green       Color = 118
	Purple      Color = 129
	Brown       Color = 130
	Lavender    Color = 147
	Red         Color = 160
	YellowGreen Color = 190
	Pink        Color = 200
	Yellow      Color = 226
)

func WithColor(color Color, msg string) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", int(color), msg)
}

type View struct {
	Width  int
	Height int
	Data   map[Coordinate]string
}

func newView(w, h int) View {
	return View{
		Width:  w,
		Height: h,
		Data:   make(map[Coordinate]string),
	}
}

func (v *View) String() string {
	b := strings.Builder{}
	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
			c := Coordinate{x, y}
			b.WriteString(v.Data[c])
		}
		b.WriteString("\n")
	}
	return b.String()
}

// ApplyWorld populates the view with the data in the world.
func (v *View) ApplyWorld(world World) {
	for c, cell := range world.Grid {
		switch cell.Type {
		case CellTypeDelta:
			v.Data[c] = WithColor(Yellow, DeltaSymbol)
		case CellTypeLambda:
			v.Data[c] = WithColor(Yellow, LambdaSymbol)
		case CellTypeSigma:
			v.Data[c] = WithColor(Yellow, SigmaSymbol)
		case CellTypeOmega:
			v.Data[c] = WithColor(Yellow, OmegaSymbol)
		}
	}
}

// ApplyBitStream populates the view with the bit stream
func (v *View) ApplyBitStream(world World) {
	for c, bs := range world.BitStream {
		v.Data[c] = bs
	}
}

// ApplyPlayer masks the bit stream
func (v *View) ApplyPlayer(player Player) {
	v.Data[player.Location] = WithColor(Purple, PlayerSymbol)

	// view distance

}
