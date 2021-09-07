package zen_doctor

import "sort"

type Player struct {
	Location      Coordinate
	Threat        float32
	Inventory     []Loot
	CurrentAction playerAction
	DataCollected map[LootType]float32
}

type ActionType int

const (
	ActionTypeNone ActionType = iota
	ActionTypeLoot
	ActionTypeExit
)

func (at ActionType) String() string {
	switch at {
	case ActionTypeLoot:
		return "Hacking"
	case ActionTypeExit:
		return "Leaving"
	}
	return ""
}

// holds data about a playerAction in progress
type playerAction struct {
	Type     ActionType
	Progress float32
	Location Coordinate
}

func (l *playerAction) tick(rate float32) {
	l.Progress += rate
	if l.Progress < 0 {
		l.Progress = 0
	}
	if l.Progress > 100 {
		l.Progress = 100
	}
}

func (l *playerAction) encounter(c Coordinate) {
	if c.X != l.Location.X || c.Y != l.Location.Y {
		l.Progress = 0
		l.Location = c
	}
}
func (l *playerAction) IsComplete() bool {
	return l.Progress >= 100
}

func (l *playerAction) IsActive() bool {
	return l.Progress > 0
}

func newPlayer(loc Coordinate) Player {
	return Player{
		Location:      loc,
		Threat:        0,
		DataCollected: map[LootType]float32{},
	}
}

func (p *Player) tickThreat(rate, max float32) {
	p.Threat += rate
	// clamp to reasonable values
	if p.Threat < 0 {
		p.Threat = 0
	}
}

func (p *Player) tickAction(t ActionType, rate float32) {
	if p.CurrentAction.Type != t {
		p.CurrentAction = playerAction{
			Type:     t,
			Progress: 0,
		}
	}
	p.CurrentAction.tick(rate)
}

func (p *Player) encounter(t ActionType, c Coordinate) {
	if p.CurrentAction.Type != t {
		p.CurrentAction = playerAction{
			Type:     t,
			Location: c,
		}
	}
	p.CurrentAction.encounter(c)
}

func (p *Player) isDetected(maxThreat float32) bool {
	return p.Threat >= maxThreat
}

func (p *Player) CollectLoot(loot Loot) {
	if loot.Type != LootTypeEmpty {
		p.Inventory = append(p.Inventory, loot)
		sort.Slice(p.Inventory, func(i, j int) bool {
			return p.Inventory[i].Rarity > p.Inventory[j].Rarity
		})
		if val, ok := p.DataCollected[loot.Type]; ok {
			p.DataCollected[loot.Type] = val + loot.Data
		} else {
			p.DataCollected[loot.Type] = loot.Data
		}
		p.CurrentAction.Progress = 0
	}
}
