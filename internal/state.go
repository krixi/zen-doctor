package zen_doctor

import "sync"

type GameState struct {
	CurrentLevel Level
	Player       Player
	World        World
	View         View
	mu           sync.Mutex
}

func NewGameState(level Level) GameState {
	l := GetLevel(level)
	return GameState{
		CurrentLevel: level,
		World:        newWorld(l),
		Player:       newPlayer(Coordinate{0, 0}),
		View:         newView(l.Width, l.Height),
	}
}

// We want to assemble a string that represents the final game state for this frame, so we do it in layers.
func (s *GameState) String() string {
	s.View.ApplyBitStream(s.World)
	s.View.ApplyWorld(s.World)
	s.View.ApplyPlayer(s)
	return s.View.String()
}

func (s *GameState) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.World.shiftBitStream(0, 1)
}

func (s *GameState) MovePlayer(dir Direction) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c := s.Player.Location
	switch dir {
	case MoveUp:
		if c.Y-1 >= 0 {
			c.Y--
		}
	case MoveDown:
		if c.Y+1 < s.World.height {
			c.Y++
		}
	case MoveLeft:
		if c.X-1 >= 0 {
			c.X--
		}
	case MoveRight:
		if c.X+1 < s.World.width {
			c.X++
		}
	}
	s.Player.Location = c
}
