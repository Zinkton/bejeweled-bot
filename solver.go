package main

import (
	"fmt"
	"math/rand"
)

type GemColor uint8

const (
	Empty GemColor = iota
	Red
	Orange
	Yellow
	Green
	Blue
	Purple
	White
	Hypercube
)

type Move struct {
	Row int
	Col int
	Dir string
}

func (g GemColor) String() string {
	switch g {
	case Empty:
		return "Empty"
	case Red:
		return "Red"
	case Orange:
		return "Orange"
	case Yellow:
		return "Yellow"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	case Purple:
		return "Purple"
	case White:
		return "White"
	case Hypercube:
		return "Hypercube"
	default:
		return "Unknown"
	}
}

func gemToAscii(g GemColor) string {
	switch g {
	case Empty:
		return " . "
	case Red:
		return "\033[31m ♦ \033[0m"
	case Orange:
		return "\033[38;5;214m ▲ \033[0m"
	case Yellow:
		return "\033[33m ★ \033[0m"
	case Green:
		return "\033[32m ■ \033[0m"
	case Blue:
		return "\033[34m ▼ \033[0m"
	case Purple:
		return "\033[35m ● \033[0m"
	case White:
		return "\033[37m ✦ \033[0m"
	default:
		return " ? "
	}
}

func printGrid(grid [8][8]GemColor) {
	fmt.Println("--- Bejeweled Grid ---")

	for _, row := range grid {
		for _, gem := range row {
			fmt.Print(gemToAscii(gem))
		}
		fmt.Println()
	}
	fmt.Println("----------------------")
}

func generateGrid() [8][8]GemColor {
	var grid [8][8]GemColor

	for r := range 8 {
		for c := range 8 {
			var newColor GemColor

			for {
				newColor = GemColor(rand.Intn(7) + 1)

				if c >= 2 && grid[r][c-1] == newColor && grid[r][c-2] == newColor {
					continue
				}

				if r >= 2 && grid[r-1][c] == newColor && grid[r-2][c] == newColor {
					continue
				}

				break
			}
			grid[r][c] = newColor
		}
	}
	return grid
}

func hasMatches(grid [8][8]GemColor) bool {
	for r := range 8 {
		for c := 0; c <= 5; c++ {
			if grid[r][c] != Empty && grid[r][c] == grid[r][c+1] && grid[r][c] == grid[r][c+2] {
				return true
			}
		}
	}

	for r := 0; r <= 5; r++ {
		for c := range 8 {
			if grid[r][c] != Empty && grid[r][c] == grid[r+1][c] && grid[r][c] == grid[r+2][c] {
				return true
			}
		}
	}

	return false
}

func checkSwap(grid *[8][8]GemColor, r1, c1, r2, c2 int) bool {
	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]

	matchFound := hasMatches(*grid)

	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]

	return matchFound
}

func findHint(grid [8][8]GemColor) (Move, bool) {
	gridPtr := &grid

	for r := range 8 {
		for c := range 8 {
			if c < 7 {
				if checkSwap(gridPtr, r, c, r, c+1) {
					return Move{Row: r, Col: c, Dir: "Right"}, true
				}
			}

			if r < 7 {
				if checkSwap(gridPtr, r, c, r+1, c) {
					return Move{Row: r, Col: c, Dir: "Down"}, true
				}
			}
		}
	}

	return Move{}, false
}
