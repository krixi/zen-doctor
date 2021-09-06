package zen_doctor

import "sort"

type Player struct {
	Location    Coordinate
	Threat      float32
	Inventory   []Loot
	CurrentLoot looting
}

// holds data about a looting in progress
type looting struct {
	Progress float32
	Location Coordinate
}

func (l *looting) tick(rate float32) {
	l.Progress += rate
	if l.Progress < 0 {
		l.Progress = 0
	}
	if l.Progress > 100 {
		l.Progress = 100
	}
}

func (l *looting) encounter(c Coordinate) {
	if c.X != l.Location.X || c.Y != l.Location.Y {
		l.Progress = 0
		l.Location = c
	}
}
func (l *looting) IsComplete() bool {
	return l.Progress >= 100
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
	}
}

func (p *Player) tickLoot(rate float32) {
	p.CurrentLoot.tick(rate)
}

func (p *Player) encounterLoot(c Coordinate) {
	p.CurrentLoot.encounter(c)
}

func (p *Player) isDetected(maxThreat float32) bool {
	return p.Threat >= maxThreat
}

func (p *Player) CollectLoot(loot Loot) {
	if loot.Type != LootTypeEmpty {
		p.Inventory = append(p.Inventory, loot)
		sort.Slice(p.Inventory, func(i, j int) bool {
			return p.Inventory[i].Value > p.Inventory[j].Value
		})
		p.CurrentLoot.Progress = 0
	}
}
