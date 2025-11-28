package lib

// Direction represents the 8 possible neighbor directions in the board graph
type Direction uint8

const (
	DirUp Direction = iota
	DirUpRight
	DirRight
	DirDownRight
	DirDown
	DirDownLeft
	DirLeft
	DirUpLeft
)

const dirCount = 8

// Opposite returns the opposite direction
func (d Direction) Opposite() Direction {
	return (d + 4) % dirCount
}

// IsVertical checks if direction is vertical (up or down)
func (d Direction) IsVertical() bool {
	return d == DirUp || d == DirDown
}

// IsHorizontal checks if direction is horizontal (left or right)
func (d Direction) IsHorizontal() bool {
	return d == DirLeft || d == DirRight
}

// IsDiagonal checks if direction is diagonal
func (d Direction) IsDiagonal() bool {
	return !d.IsVertical() && !d.IsHorizontal()
}
