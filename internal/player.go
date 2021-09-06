package zen_doctor

import (
	"strings"
)

type Player struct {
	Location  Coordinate
	Threat    float32
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

func (p *Player) tickThreat(rate float32) {
	p.Threat += rate
}

// ThreatMeter returns the formatted display string for the player's current threat.
func (p *Player) ThreatMeter(size int) string {
	b := strings.Builder{}
	threat := int(p.Threat)
	for i := 0; i < threat; i++ {
		var color Color
		if threat < size/3 {
			color = Green
		} else if threat < (size/3)+(size/3) {
			color = Yellow
		} else {
			color = Red
		}
		b.WriteString(WithColor(color, FullBlockSymbol))
	}
	return b.String()
}

func (p *Player) isDetected(maxThreat int) bool {
	return int(p.Threat) >= maxThreat
}
