package game

import (
	"math/rand"
)

// Game represents a Tetris game instance
type Game struct {
	Board *Board
}

// NewGame creates a new game instance
func NewGame() *Game {
	game := &Game{
		Board: NewBoard(),
	}
	game.SpawnPiece()
	return game
}

// SpawnPiece creates a new piece at the top of the board
func (g *Game) SpawnPiece() {
	if g.Board.GameOver {
		return
	}

	// Choose a random piece
	piece := Pieces[rand.Intn(len(Pieces))]
	piece.X = BoardWidth/2 - len(piece.Shape[0])/2
	piece.Y = 0

	// Check if the new piece can be placed
	if g.checkCollision(&piece) {
		g.Board.GameOver = true
		return
	}

	g.Board.CurrentPiece = &piece
}

// MovePiece moves the current piece in the specified direction
func (g *Game) MovePiece(dx, dy int) bool {
	if g.Board.CurrentPiece == nil || g.Board.GameOver {
		return false
	}

	piece := *g.Board.CurrentPiece
	piece.X += dx
	piece.Y += dy

	if g.checkCollision(&piece) {
		if dy > 0 {
			// Piece has landed
			g.lockPiece()
			g.clearLines()
			g.SpawnPiece()
		}
		return false
	}

	g.Board.CurrentPiece = &piece
	return true
}

// RotatePiece rotates the current piece clockwise
func (g *Game) RotatePiece() bool {
	if g.Board.CurrentPiece == nil || g.Board.GameOver {
		return false
	}

	piece := *g.Board.CurrentPiece
	// Create new rotated shape
	oldShape := piece.Shape
	newShape := make([][]bool, len(oldShape[0]))
	for i := range newShape {
		newShape[i] = make([]bool, len(oldShape))
		for j := range newShape[i] {
			newShape[i][j] = oldShape[len(oldShape)-1-j][i]
		}
	}
	piece.Shape = newShape

	if g.checkCollision(&piece) {
		return false
	}

	g.Board.CurrentPiece = &piece
	return true
}

// checkCollision checks if a piece collides with the board boundaries or other pieces
func (g *Game) checkCollision(piece *Piece) bool {
	for y := range piece.Shape {
		for x := range piece.Shape[y] {
			if !piece.Shape[y][x] {
				continue
			}

			newY := piece.Y + y
			newX := piece.X + x

			// Check boundaries
			if newX < 0 || newX >= BoardWidth || newY >= BoardHeight {
				return true
			}

			// Check collision with locked pieces
			if newY >= 0 && g.Board.Grid[newY][newX] == "█" {
				return true
			}
		}
	}
	return false
}

// lockPiece locks the current piece in place
func (g *Game) lockPiece() {
	if g.Board.CurrentPiece == nil {
		return
	}

	piece := g.Board.CurrentPiece
	for y := range piece.Shape {
		for x := range piece.Shape[y] {
			if piece.Shape[y][x] {
				newY := piece.Y + y
				newX := piece.X + x
				if newY >= 0 && newY < BoardHeight && newX >= 0 && newX < BoardWidth {
					g.Board.Grid[newY][newX] = "█"
				}
			}
		}
	}
}

// clearLines removes completed lines and updates the score
func (g *Game) clearLines() {
	linesCleared := 0
	for y := BoardHeight - 1; y >= 0; y-- {
		if g.isLineFull(y) {
			g.removeLine(y)
			linesCleared++
			y++ // Check the same line again after shifting
		}
	}

	// Update score (more points for clearing multiple lines at once)
	if linesCleared > 0 {
		g.Board.Score += linesCleared * linesCleared * 100
	}
}

// isLineFull checks if a line is complete
func (g *Game) isLineFull(y int) bool {
	for x := 0; x < BoardWidth; x++ {
		if g.Board.Grid[y][x] != "█" {
			return false
		}
	}
	return true
}

// removeLine removes a line and shifts everything above down
func (g *Game) removeLine(y int) {
	// Shift everything down
	for i := y; i > 0; i-- {
		copy(g.Board.Grid[i], g.Board.Grid[i-1])
	}

	// Clear top line
	for x := 0; x < BoardWidth; x++ {
		g.Board.Grid[0][x] = "·"
	}
}
