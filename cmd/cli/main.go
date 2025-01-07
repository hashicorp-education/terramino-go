package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp-education/terraminogo/internal/game"
)

const (
	clearScreen = "\033[H\033[2J"
	moveCursor  = "\033[%d;%dH"
	escapeKey   = 27
	spaceKey    = 32
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Use alternate screen buffer and hide cursor
	fmt.Print("\033[?1049h")
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?1049l\033[?25h")

	clearTerminal()
	fmt.Print("Terramino CLI\r\n")
	fmt.Print("Controls:\r\n")
	fmt.Print("← →   : Move left/right\r\n")
	fmt.Print("↑     : Rotate\r\n")
	fmt.Print("↓     : Soft drop\r\n")
	fmt.Print("Space : Hard drop\r\n")
	fmt.Print("q     : Quit\r\n")
	fmt.Print("\r\nPress Enter to start...\r\n")

	reader.ReadString('\n')

	// Now set up raw mode after user presses Enter
	rawMode := exec.Command("stty", "raw", "-echo")
	rawMode.Stdin = os.Stdin
	_ = rawMode.Run()

	// Restore on exit
	defer func() {
		cookedMode := exec.Command("stty", "-raw", "echo")
		cookedMode.Stdin = os.Stdin
		_ = cookedMode.Run()
	}()

	startGame(reader)
}

func startGame(reader *bufio.Reader) {
	g := game.NewGame()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Start input reading in a separate goroutine
	inputChan := make(chan string)
	go func() {
		for {
			char, err := reader.ReadByte()
			if err != nil {
				continue
			}

			if char == escapeKey {
				// Read the next two bytes for arrow keys
				secondChar, err := reader.ReadByte()
				if err != nil {
					continue
				}
				if secondChar == '[' {
					thirdChar, err := reader.ReadByte()
					if err != nil {
						continue
					}
					switch thirdChar {
					case 'A': // Up arrow
						inputChan <- "up"
					case 'B': // Down arrow
						inputChan <- "down"
					case 'C': // Right arrow
						inputChan <- "right"
					case 'D': // Left arrow
						inputChan <- "left"
					}
				}
				continue
			}

			if char == spaceKey {
				inputChan <- "space"
			} else if char == 'q' {
				inputChan <- "quit"
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			g.MovePiece(0, 1) // Move down automatically
			renderGame(g)
		case input := <-inputChan:
			switch input {
			case "quit":
				return
			case "left":
				g.MovePiece(-1, 0)
			case "right":
				g.MovePiece(1, 0)
			case "up":
				g.RotatePiece()
			case "down":
				g.MovePiece(0, 1)
			case "space":
				// Hard drop - move down until collision
				for g.MovePiece(0, 1) {
					// Keep moving down
				}
			}
			renderGame(g)
		}

		if g.Board.GameOver {
			fmt.Printf("\nGame Over! Score: %d\n", g.Board.Score)
			fmt.Print("Press 'q' to quit...\n")
			for {
				char, _ := reader.ReadByte()
				if char == 'q' {
					return
				}
			}
		}
	}
}

func renderGame(g *game.Game) {
	clearTerminal()
	state := g.Board.GetState()

	// Print title
	fmt.Print("Terramino\r\n")

	// Print top border
	fmt.Print("┌" + strings.Repeat("──", game.BoardWidth) + "┐\r\n")

	// Print board
	for _, row := range state.Grid {
		fmt.Print("│")
		for _, cell := range row {
			fmt.Print(cell + cell) // Print each cell twice for better aspect ratio
		}
		fmt.Print("│\r\n")
	}

	// Print bottom border
	fmt.Print("└" + strings.Repeat("──", game.BoardWidth) + "┘\r\n")
	fmt.Printf("Score: %d\r\n", state.Score)
	fmt.Print("Press 'q' to quit...\r\n")
}

func clearTerminal() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print(clearScreen)
	}
}
