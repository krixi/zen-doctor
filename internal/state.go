package zen_doctor

import (
	"sync"
)

type GameState struct {
	level  LevelConfig
	player Player
	world  World
	view   View
	mu     sync.Mutex
}

func NewGameState(level Level) GameState {
	l := GetLevel(level)
	return GameState{
		level:  l,
		world:  newWorld(l),
		player: newPlayer(Coordinate{0, 0}),
		view:   newView(l.Width, l.Height),
	}
}

func (s *GameState) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.view.Apply(s)
	return s.view.String()
}

func (s *GameState) ThreatMeter() string {
	return s.view.ThreatMeter(s.player.Threat, s.Level().MaxThreat)
}

func (s *GameState) LootProgressMeter() string {
	return s.view.LootProgressMeter(s.player.CurrentLoot.Progress, 100)
}

func (s *GameState) TickWorld() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.world.TickLoot()
}

func (s *GameState) TickBitStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.world.shiftBitStream(MoveDown)
	s.tickCollisions()
}

func (s *GameState) TickPlayer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// handle player looting
	if s.world.DidCollideWithLoot(s.player.Location) {
		s.player.encounterLoot(s.player.Location)
		s.player.tickLoot(s.level.LootSpeed)

		// move loot to inventory once it's completely looted.
		if s.player.CurrentLoot.IsComplete() {
			s.player.CollectLoot(s.world.ExtractLoot(s.player.Location))
		}
	} else {
		s.player.tickLoot(s.level.LootSpeedDecay)

		// don't decay threat while looting
		s.player.tickThreat(s.level.ThreatDecay, s.level.MaxThreat)
	}
}

// check for collisions with bad bits
func (s *GameState) tickCollisions() {
	// note: not wrapped in mutex since this is called from mutex protected calls already.
	if threat, ok := s.world.DidCollideWith(s.player.Location, RevealedBitHelpful); ok {
		// good bits
		s.player.tickThreat(-1*threat, s.level.MaxThreat)
		s.world.NeutralizeBit(s.player.Location)
	}

	if threat, ok := s.world.DidCollideWith(s.player.Location, RevealedBitHarmful); ok {
		// bad bits
		s.player.tickThreat(threat, s.level.MaxThreat)
	}
}

func (s *GameState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.player.Threat = 0
	s.player.Location = Coordinate{0, 0}
}

func (s *GameState) IsGameOver() bool {
	return s.player.isDetected(s.level.MaxThreat)
}

func (s *GameState) Level() LevelConfig {
	return s.level
}

func (s *GameState) Inventory() []Loot {
	return s.player.Inventory
}

func (s *GameState) DataWanted() string {
	return s.view.DataWanted(s)
}

func (s *GameState) DataCollected() string {
	return s.view.DataCollected(s)
}

func (s *GameState) MovePlayer(dir Direction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.player.Location
	switch dir {
	case MoveUp:
		if c.Y-1 >= 0 {
			c.Y--
		}
	case MoveDown:
		if c.Y+1 < s.level.Height-1 {
			c.Y++
		}
	case MoveLeft:
		if c.X-1 >= 0 {
			c.X--
		}
	case MoveRight:
		if c.X+1 < s.level.Width-1 {
			c.X++
		}
	}
	s.player.Location = c
	level := s.Level()
	s.player.tickThreat(level.MovementThreat, level.MaxThreat)
	s.tickCollisions()
}
