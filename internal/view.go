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
	Blue        Color = 20
	LightBlue   Color = 81
	Teal        Color = 85
	Green       Color = 118
	Purple      Color = 129
	Brown       Color = 130
	Lavender    Color = 147
	Red         Color = 1
	YellowGreen Color = 190
	Pink        Color = 200
	Orange      Color = 208
	Yellow      Color = 226
	DarkGray    Color = 235
	LightGray   Color = 245
	White       Color = 255
)

func GameOver(didWin bool) string {
	if didWin {
		return WithColor(Green, "You did it!")
	}
	return WithColor(Red, "Your hack has been detected!")
}

func WithColor(color Color, msg string) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", int(color), msg)
}

func WithBackground(color Color, msg string) string {
	return fmt.Sprintf("\x1b[48;5;%dm%s\x1b[0m", int(color), msg)
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
func (v *View) applyWorld(world *World) {
	for c, loot := range world.Loot {
		switch loot.Type {
		case LootTypeEmpty:
			continue
		default:
			v.Data[c] = loot.WithIntegrity(QuestionSymbol)
		}
	}

	if world.Exit != nil {
		v.Data[*world.Exit] = WithColor(Pink, ReferenceMarkSymbol)
	}
}

// applyBitStream populates the view with the bit stream
func (v *View) applyBitStream(world *World) {
	for c, bs := range world.BitStream {
		v.Data[c] = WithColor(DarkGray, bs.ViewHidden())
	}
}

// Apply updates the view from the state.
// We want to assemble a string that represents the final game state for this frame, so we do it in layers.
func (v *View) Apply(s *GameState) {
	// First is the bit stream
	v.applyBitStream(&s.world)

	// Then is the world.
	v.applyWorld(&s.world)

	// Then finally, the player
	c := s.player.Location
	v.Data[c] = WithColor(YellowGreen, PlayerSymbol)

	// mask for view distance
	level := s.Level()
	vdx := level.ViewDistX
	vdy := level.ViewDistY
	for x := c.X - vdx; x <= c.X+vdx; x++ {
		for y := c.Y - vdy; y <= c.Y+vdy; y++ {

			offset := Coordinate{x, y}

			// skip out of bounds
			if _, ok := v.Data[offset]; !ok {
				continue
			}

			// give player a dark background but that's it
			v.Data[offset] = WithBackground(DarkGray, v.Data[offset])
			if x == c.X && y == c.Y {
				continue
			}
			loot := s.world.Loot[offset]
			if loot.Type != LootTypeEmpty {
				v.Data[offset] = WithBackground(DarkGray, loot.String())
				continue
			}

			// show the bit stream around them with a background and a color based on whether it's good or bad
			bs := s.world.BitStream[offset]
			if bs.Hidden != BitTypeEmpty {
				color := LightGray
				switch bs.Revealed {
				case RevealedBitHelpful:
					color = Green
				case RevealedBitHarmful:
					color = Red
				}
				v.Data[offset] = WithBackground(DarkGray, WithColor(color, bs.ViewRevealed()))
			}
		}
	}
}

// ThreatMeter scales the string to fit inside the view correctly.
// it assumes the meter is always the width of the view.
func (v *View) ThreatMeter(current, max float32) string {
	b := strings.Builder{}

	// find the percent, convert that to an int over v.Width
	percent := current / max
	threat := int(percent * float32(v.Width))

	for i := 0; i < threat; i++ {
		var color Color
		if threat < v.Width/3 {
			color = Green
		} else if threat < (v.Width/3)+(v.Width/3) {
			color = Yellow
		} else {
			color = Red
		}
		b.WriteString(WithColor(color, FullBlockSymbol))
	}
	return b.String()
}

func (v *View) ActionProgressMeter(current, max float32) string {
	b := strings.Builder{}
	// find the percent, convert that to an int over v.Width
	if current > max {
		current = max
	}
	percent := current / max
	progress := int(percent * float32(v.Width))
	for i := 0; i < progress; i++ {
		b.WriteString(WithColor(LightBlue, FullBlockSymbol))
	}
	return b.String()
}

func (v *View) DataWanted(state *GameState) string {
	b := strings.Builder{}
	for _, want := range state.Level().WinConditions {
		b.WriteString(fmt.Sprintf("%s %.0f\n", want.Type.String(), want.Amount))
	}
	return b.String()
}

func (v *View) DataCollected(state *GameState) string {
	b := strings.Builder{}
	for _, want := range state.level.WinConditions {
		if amount, ok := state.player.DataCollected[want.Type]; ok {
			str := fmt.Sprintf("%s %.0f\n", want.Type.String(), amount)
			if amount > want.Amount {
				str = WithColor(Green, str)
			}
			b.WriteString(str)
		}
	}
	for _, want := range state.level.Bonus {
		if amount, ok := state.player.DataCollected[want]; ok {
			b.WriteString(fmt.Sprintf("%s %.0f\n", want.String(), amount))
		}
	}
	if state.isExitUnlocked() {
		b.WriteString(fmt.Sprintf("Exit %s unlocked!\n", WithColor(Pink, ReferenceMarkSymbol)))
	}
	return b.String()
}
