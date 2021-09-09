package zen_doctor

import (
	"fmt"
	"strings"
	"time"
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
	Black       Color = 232
	DarkGray    Color = 235
	LightGray   Color = 245
	White       Color = 255
)

func GameOver(didWin bool, elapsed time.Duration, mode CompatibilityMode, collection ...Loot) string {

	// group collection by type and then by rarity. We want a display like:
	// 1Δ 5Δ 13Δ 21Δ 6Δ
	byType := make(map[LootType]map[Rarity]int)
	for _, loot := range collection {
		lt := loot.Type
		r := loot.Rarity
		if _, ok := byType[lt]; !ok {
			byType[lt] = make(map[Rarity]int)
		}
		if count, ok := byType[lt][r]; ok {
			byType[lt][r] = count + 1
		} else {
			byType[lt][r] = 1
		}
	}

	getLine := func(lt LootType, counts map[Rarity]int) (string, int) {
		b := strings.Builder{}
		hierarchy := []Rarity{Legendary, Epic, Rare, Uncommon, Common, Junk}
		found := 0
		for _, rarity := range hierarchy {
			if count, ok := counts[rarity]; ok {
				b.WriteString(fmt.Sprintf("%d%s ", count, WithColor(rarity.Color(), lt.SymbolForMode(mode))))
				found++
			}
		}
		if found > 0 {
			b.WriteString("\n")
		}
		return b.String(), found
	}

	b := strings.Builder{}
	if didWin {
		b.WriteString(WithColor(Green, "You did it! Results:\n"))
	} else {
		b.WriteString(WithColor(Red, "You were caught! Results:\n"))
	}
	hierarchy := []LootType{LootTypeDelta, LootTypeLambda, LootTypeSigma, LootTypeOmega}
	for _, lt := range hierarchy {
		if counts, ok := byType[lt]; ok {
			line, found := getLine(lt, counts)
			if found > 0 {
				b.WriteString(line)
			}
		}
	}
	b.WriteString(ElapsedTime(elapsed))
	b.WriteString("\nPress <space> to retry\n")
	return b.String()
}

func ElapsedTime(elapsed time.Duration) string {
	d := elapsed.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	} else if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%2.1fs", elapsed.Seconds())
}

func WithColor(color Color, msg string) string {
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", int(color), msg)
}

func WithBackground(color Color, msg string) string {
	return fmt.Sprintf("\x1b[48;5;%dm%s\x1b[0m", int(color), msg)
}

type View struct {
	Width      int
	Height     int
	Mode       CompatibilityMode
	ExitSymbol AnimatedSymbol
	Data       map[Coordinate]Cell
}

type Cell struct {
	Background Color
	Foreground Color
	Symbol     string
}

func (c Cell) String() string {
	return WithBackground(c.Background, WithColor(c.Foreground, c.Symbol))
}

func newView(w, h int, mode CompatibilityMode) View {
	return View{
		Width:      w,
		Height:     h,
		Mode:       mode,
		ExitSymbol: &AnimatedExit,
		Data:       make(map[Coordinate]Cell),
	}
}

func (v *View) SetCompatibility(mode CompatibilityMode) {
	v.Mode = mode
}

func (v *View) String() string {
	b := strings.Builder{}
	for y := 0; y < v.Height; y++ {
		for x := 0; x < v.Width; x++ {
			c := Coordinate{x, y}
			b.WriteString(v.Data[c].String())
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Apply updates the view from the state.
// When we render, we need to have a string that represents the X,Y plane for the game. We can do this in a single
// pass, and have a predictable draw order, by assembling the three components of a cell (background color, foreground color,
// and symbol), and overwriting these with a draw order priority. Things at the top of the for-loop are lower priority,
// with the highest priority at the bottom.
func (v *View) Apply(s *GameState) {
	for x := 0; x < s.level.Width; x++ {
		for y := 0; y < s.level.Height; y++ {
			c := Coordinate{x, y}

			// Hidden bit stream is always shown if there's nothing else.
			bs := s.bits.stream[c]
			cell := Cell{
				Background: Black,
				Foreground: DarkGray,
				Symbol:     bs.ViewHidden(),
			}

			// footprints
			if footprint, ok := s.world.Footprints[c]; ok && bs.Hidden == BitTypeEmpty {
				cell.Foreground, cell.Symbol = footprint.WithIntensity()
			}

			// determine if the bit stream and loot should be revealed, by checking the distance from this tile to the player
			inPlayerRange := s.player.Location.InRange(s.level.ViewDist, c)

			// revealed bit stream
			if inPlayerRange {
				cell.Background = DarkGray
				if bs.Hidden != BitTypeEmpty {
					cell.Foreground, cell.Symbol = bs.ViewRevealed(v.Mode)
				}
			}

			// loot
			if loot, ok := s.world.Loot[c]; ok {
				if loot.Type != LootTypeEmpty {
					if inPlayerRange {
						cell.Foreground, cell.Symbol = loot.SymbolForMode(v.Mode)
					} else {
						cell.Foreground, cell.Symbol = loot.WithIntegrity(QuestionSymbol)
					}
				}
			}

			// Exit location
			if s.world.Exit != nil && c.Equals(*s.world.Exit) {
				cell.Foreground, cell.Symbol = v.exitSymbol()
			}

			// finally, player location
			if c.Equals(s.player.Location) {
				cell.Foreground, cell.Symbol = YellowGreen, PlayerSymbol.ForMode(v.Mode)
			}

			v.Data[c] = cell
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
		} else if threat < (2*v.Width)/3 {
			color = Yellow
		} else {
			color = Red
		}
		b.WriteString(WithColor(color, ProgressBarSymbol.ForMode(v.Mode)))
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
		b.WriteString(WithColor(LightBlue, ProgressBarSymbol.ForMode(v.Mode)))
	}
	return b.String()
}

func (v *View) DataWanted(state *GameState) string {
	b := strings.Builder{}
	for _, want := range state.Level().WinConditions {
		b.WriteString(fmt.Sprintf("%s %.0f\n", want.Type.SymbolForMode(v.Mode), want.Amount))
	}
	return b.String()
}

func (v *View) DataCollected(state *GameState) string {
	b := strings.Builder{}
	for _, want := range state.level.WinConditions {
		if amount, ok := state.player.DataCollected[want.Type]; ok {
			str := fmt.Sprintf("%s %.0f\n", want.Type.SymbolForMode(v.Mode), amount)
			if amount > want.Amount {
				str = WithColor(Green, str)
			}
			b.WriteString(str)
		}
	}
	for _, want := range state.level.Bonus {
		if amount, ok := state.player.DataCollected[want]; ok {
			b.WriteString(fmt.Sprintf("%s %.0f\n", want.SymbolForMode(v.Mode), amount))
		}
	}
	if state.isExitUnlocked() {
		b.WriteString(fmt.Sprintf("Exit %s unlocked!\n", WithColor(v.exitSymbol())))
	}
	return b.String()
}

func (v *View) TickAnimations() {
	v.ExitSymbol.Tick()
}

func (v *View) exitSymbol() (Color, string) {
	return Pink, v.ExitSymbol.ForMode(v.Mode)
}
