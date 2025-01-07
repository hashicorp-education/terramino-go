package game

import (
	"encoding/json"
)

const (
	BoardWidth  = 10
	BoardHeight = 20
)

// Piece represents a Tetris piece
type Piece struct {
	Shape     [][]bool
	X, Y      int
	PieceType string
}

// Board represents the game board
type Board struct {
	Grid         [][]string
	CurrentPiece *Piece
	Score        int
	GameOver     bool
}

// NewBoard creates a new game board
func NewBoard() *Board {
	grid := make([][]string, BoardHeight)
	for i := range grid {
		grid[i] = make([]string, BoardWidth)
		for j := range grid[i] {
			grid[i][j] = "·"
		}
	}
	return &Board{
		Grid:  grid,
		Score: 0,
	}
}

// State represents the game state for API responses
type State struct {
	Grid     [][]string `json:"grid"`
	Score    int        `json:"score"`
	GameOver bool       `json:"gameOver"`
}

// GetState returns the current game state
func (b *Board) GetState() State {
	// Create a copy of the grid
	gridCopy := make([][]string, len(b.Grid))
	for i := range b.Grid {
		gridCopy[i] = make([]string, len(b.Grid[i]))
		copy(gridCopy[i], b.Grid[i])
	}

	// Add current piece to the grid copy
	if b.CurrentPiece != nil {
		for y := range b.CurrentPiece.Shape {
			for x := range b.CurrentPiece.Shape[y] {
				if b.CurrentPiece.Shape[y][x] {
					newY := b.CurrentPiece.Y + y
					newX := b.CurrentPiece.X + x
					if newY >= 0 && newY < BoardHeight && newX >= 0 && newX < BoardWidth {
						gridCopy[newY][newX] = "█"
					}
				}
			}
		}
	}

	return State{
		Grid:     gridCopy,
		Score:    b.Score,
		GameOver: b.GameOver,
	}
}

// ToJSON converts the game state to JSON
func (s State) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// Pieces definitions
var Pieces = []Piece{
	{
		// I piece
		Shape: [][]bool{
			{true, true, true, true},
		},
		PieceType: "I",
	},
	{
		// O piece
		Shape: [][]bool{
			{true, true},
			{true, true},
		},
		PieceType: "O",
	},
	{
		// T piece
		Shape: [][]bool{
			{false, true, false},
			{true, true, true},
		},
		PieceType: "T",
	},
	{
		// L piece
		Shape: [][]bool{
			{true, false},
			{true, false},
			{true, true},
		},
		PieceType: "L",
	},
	{
		// J piece
		Shape: [][]bool{
			{false, true},
			{false, true},
			{true, true},
		},
		PieceType: "J",
	},
	{
		// S piece
		Shape: [][]bool{
			{false, true, true},
			{true, true, false},
		},
		PieceType: "S",
	},
	{
		// Z piece
		Shape: [][]bool{
			{true, true, false},
			{false, true, true},
		},
		PieceType: "Z",
	},
}
