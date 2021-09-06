package zen_doctor

import "sync"

type GameState struct {
	currentLevel Level
	player       Player
	world        World
	view         View
	mu           sync.Mutex
}

func NewGameState(level Level) GameState {
	l := GetLevel(level)
	return GameState{
		currentLevel: level,
		world:        newWorld(l),
		player:       newPlayer(Coordinate{0, 0}),
		view:         newView(l.Width, l.Height),
	}
}

// We want to assemble a string that represents the final game state for this frame, so we do it in layers.
func (s *GameState) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.view.Apply(s)
	return s.view.String()
}

func (s *GameState) ThreatMeter() string {
	return s.player.ThreatMeter(s.GetLevel().MaxThreat)
}

func (s *GameState) TickBitStream() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.world.shiftBitStream(0, 1)
}

func (s *GameState) TickPlayer() {
	s.mu.Lock()
	defer s.mu.Unlock()

	level := GetLevel(s.currentLevel)
	s.player.tickThreat(level.ThreatRate)
}

func (s *GameState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.player.Threat = 0
	s.player.Location = Coordinate{0, 0}
}

func (s *GameState) IsGameOver() bool {
	level := GetLevel(s.currentLevel)
	return s.player.isDetected(level.MaxThreat)
}

func (s *GameState) GetLevel() LevelSettings {
	return GetLevel(s.currentLevel)
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
		if c.Y+1 < s.world.height-1 {
			c.Y++
		}
	case MoveLeft:
		if c.X-1 >= 0 {
			c.X--
		}
	case MoveRight:
		if c.X+1 < s.world.width-1 {
			c.X++
		}
	}
	s.player.Location = c
}
