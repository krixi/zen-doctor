package zen_doctor

import (
	"math/rand"
	"sync"
)

type GameState struct {
	level    LevelConfig
	player   Player
	world    World
	view     View
	mu       sync.Mutex
	complete bool
}

func NewGameState(level Level, mode CompatibilityMode) GameState {
	l := GetLevel(level)

	return GameState{
		level:  l,
		world:  newWorld(l),
		player: newPlayer(Coordinate{1+rand.Intn(l.Width-2), 1+rand.Intn(l.Height-2)}),
		view:   newView(l.Width, l.Height, mode),
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

func (s *GameState) ProgressBar() string {
	return s.view.ActionProgressMeter(s.player.CurrentAction.Progress, 100)
}

func (s *GameState) ProgressBarType() string {
	if s.player.CurrentAction.IsActive() {
		return s.player.CurrentAction.Type.String()
	}
	return ""
}

func (s *GameState) isExitUnlocked() bool {
	for _, want := range s.level.WinConditions {
		if !want.IsMet(s) {
			return false
		}
	}
	return true
}

func (s *GameState) TickWorld() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.world.TickLoot()

	// check if world exit is unlocked
	if s.isExitUnlocked() {
		s.world.UnlockExit()
	}

	s.world.TickFootprints()
}

func (s *GameState) TickAnimations() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.view.tickAnimations()
	s.world.tickAnimations()
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

	// handle player actions
	if s.world.DidCollideWithExit(s.player.Location) {
		s.player.encounter(ActionTypeExit, s.player.Location)
		s.player.tickAction(ActionTypeExit, s.level.LeaveSpeed)

		if s.player.CurrentAction.IsComplete() {
			s.complete = true
		}
	} else if s.world.DidCollideWithLoot(s.player.Location) {
		s.player.encounter(ActionTypeLoot, s.player.Location)
		s.player.tickAction(ActionTypeLoot, s.level.LootSpeed)
		// prevent loot from despawning while we loot it
		s.world.TickLootAt(s.player.Location, -2*s.level.LootDecayRate)

		// move loot to inventory once it's completely looted.
		if s.player.CurrentAction.IsComplete() {
			s.player.CollectLoot(s.world.ExtractLoot(s.player.Location))
		}
	} else {
		switch s.player.CurrentAction.Type {
		case ActionTypeLoot:
			s.player.tickAction(ActionTypeLoot, s.level.LootSpeedDecay)
		case ActionTypeExit:
			s.player.tickAction(ActionTypeExit, s.level.LeaveSpeedDecay)
		}

		// only decay threat while not performing an action
		s.player.tickThreat(s.level.ThreatDecay, s.level.MaxThreat)
	}
}

// check for collisions with bad bits
func (s *GameState) tickCollisions() {
	// note: not wrapped in mutex since this is called from mutex protected calls already.
	if threat, ok := s.world.DidCollideWithBit(s.player.Location, RevealedBitHelpful); ok {
		// good bits
		s.player.tickThreat(-1*threat, s.level.MaxThreat)
		s.world.NeutralizeBit(s.player.Location)
	}

	if threat, ok := s.world.DidCollideWithBit(s.player.Location, RevealedBitHarmful); ok {
		// bad bits
		s.player.tickThreat(threat, s.level.MaxThreat)
	}
}

func (s *GameState) IsGameOver() bool {
	return s.player.isDetected(s.level.MaxThreat)
}

func (s *GameState) IsComplete() bool {
	return s.complete
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
	s.world.Visited(c)
}
