package zen_doctor

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

func (p *Player) tickThreat(rate, max float32) {
	p.Threat += rate
	// clamp to reasonable values
	if p.Threat < 0 {
		p.Threat = 0
	} else if p.Threat > max {
		p.Threat = max
	}
}

func (p *Player) isDetected(maxThreat int) bool {
	return int(p.Threat) >= maxThreat
}
