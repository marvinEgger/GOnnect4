package lib

// Node represents a cell in the board as a graph node
type Node struct {
	Row       int
	Col       int
	Owner     Cell
	Neighbors [dirCount]*Node
}

// GetNeighbor returns the neighbor in the given direction, or nil if none
func (n *Node) GetNeighbor(dir Direction) *Node {
	return n.Neighbors[dir]
}

// TODO:  SetNeighbor sets the neighbor in the given direction
func (n *Node) SetNeighbor(dir Direction, neighbor *Node) {
}

// IsEmpty checks if the node has no owner
func (n *Node) IsEmpty() bool {
	return n.Owner == CellEmpty
}

// SetOwner sets the owner of this node
func (n *Node) SetOwner(player Cell) {
	n.Owner = player
}

// TODO: CountSequence counts consecutive nodes with same owner in given direction
func (n *Node) CountSequence(dir Direction) int {
	return -1
}

// TODO: CheckWin checks if placing a token at this node creates a winning sequence
func (n *Node) CheckWin(winLength int) bool {
	return false
}
