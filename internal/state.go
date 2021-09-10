package zen_doctor

import (
	"math/rand"
	"sync"
)

type GameState struct {
	level    *LevelConfig
	player   Player
	bits     BitStream
	world    World
	view     View
	mu       sync.Mutex
	complete bool
}

func NewGameState(level Level, mode CompatibilityMode) GameState {
	l := GetLevel(level)
	return NewGameStateWithPlayerAt(Coordinate{
		X: 1 + rand.Intn(l.Width-2),
		Y: 1 + rand.Intn(l.Height-2),
	}, level, mode)
}

func NewGameStateWithPlayerAt(c Coordinate, level Level, mode CompatibilityMode) GameState {
	l := GetLevel(level)
	return GameState{
		level:  &l,
		bits:   newBitStream(&l),
		player: newPlayer(c),
		view:   newView(l.Width, l.Height, mode),
		world:  newWorld(&l),
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

	s.view.TickAnimations()
	s.bits.TickAnimations()
}

func (s *GameState) TickBitStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.level.Updater.Tick(&s.bits)
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
		s.world.TickLootAt(s.player.Location, -2*s.level.DataDecayRate)

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
		s.player.tickThreat(s.level.ThreatDecay)
	}
}

// check for collisions with bad bits
func (s *GameState) tickCollisions() {
	// note: not wrapped in mutex since this is called from mutex protected calls already.
	if threat, ok := s.bits.DidCollideWithBit(s.level, s.player.Location, RevealedBitHelpful); ok {
		// good stream
		s.player.tickThreat(-1 * threat)
		s.bits.NeutralizeBit(s.player.Location)
	}

	if threat, ok := s.bits.DidCollideWithBit(s.level, s.player.Location, RevealedBitHarmful); ok {
		// bad stream
		s.player.tickThreat(threat)
	}
}

func (s *GameState) IsGameOver() bool {
	return s.player.isDetected(s.level.MaxThreat)
}

func (s *GameState) IsComplete() bool {
	return s.complete
}

func (s *GameState) PlayerLocation() Coordinate {
	return s.player.Location
}

func (s *GameState) Level() *LevelConfig {
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

	c := s.player.HandleMoveInput(dir, s.level.Width, s.level.Height)
	s.player.tickThreat(s.level.MovementThreat)
	s.tickCollisions()
	s.world.Visited(c)
}

func (s *GameState) TickMovement() {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, applyDmg := s.player.tickMove(s.level.Width, s.level.Height, s.level.MovementThreat)
	if applyDmg {
		s.tickCollisions()
	}
	s.world.Visited(c)
}
