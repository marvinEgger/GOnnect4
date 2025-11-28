package lib

const (
	Rows      = 6
	Cols      = 7
	WinLength = 4
)

// Board represents the game board as a graph of connected nodes
type Board struct {
	nodes      [][]*Node
	colHeights [Cols]int
	rows       int
	cols       int
}

// TODO: NewBoard creates a new board and builds the node graph
func NewBoard() *Board {
	return &Board{}
}

// TODO: buildGraph creates all nodes and establishes neighbor relationships
func (b *Board) buildGraph() {

}

// TODO: CanPlay checks if a column can accept a token
func (b *Board) CanPlay(col int) bool {
	return false
}

// TODO: Play drops a token in the given column for the given player
func (b *Board) Play(col int, player Cell) (*Node, bool) {
	return nil, false
}

// TODO: CheckWin checks if the last played node creates a winning condition
func (b *Board) CheckWin(node *Node) bool {
	return false
}

// TODO: IsFull checks if the board is completely full
func (b *Board) IsFull() bool {
	return false
}

// TODO: GetNode returns the node at given position
func (b *Board) GetNode(row, col int) *Node {
	return nil
}

// TODO: GetLastPlayedNode returns the node at the top of a column
func (b *Board) GetLastPlayedNode(col int) *Node {
	return nil
}

// TODO: Reset clears the board for a new game
func (b *Board) Reset() {

}

// TODO: ToArray exports the board state as a 2D array
func (b *Board) ToArray() [Rows][Cols]Cell {
	return [Rows][Cols]Cell{}
}
