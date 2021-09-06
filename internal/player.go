package zen_doctor

type Player struct {
	Location  Coordinate
	Threat    int
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
