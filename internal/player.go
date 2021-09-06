package zen_doctor

type Player struct {
	Location Coordinate
	Threat   float32
}

func newPlayer(loc Coordinate) Player {
	return Player{
		Location: loc,
		Threat:   0,
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

func (p *Player) isDetected(maxThreat float32) bool {
	return p.Threat >= maxThreat
}
