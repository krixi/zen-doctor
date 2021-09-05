package zen_doctor

import "fmt"

func CalculateViewPosition(width, height, screenWidth, screenHeight int) (int, int) {
	x := screenWidth/2 - width/2
	y := screenHeight/2 - height/2
	return x, y
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
