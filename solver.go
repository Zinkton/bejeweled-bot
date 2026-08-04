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
)

type Move struct {
	Row      int
	Col      int
	Dir      string
	Priority int
}

func (g GemColor) String() string {
	names := []string{"Empty", "Red", "Orange", "Yellow", "Green", "Blue", "Purple", "White"}
	if int(g) < len(names) {
		return names[g]
	}
	return "Unknown"
}

func gemToAscii(g GemColor) string {
	ascii := []string{
		" . ",
		"\033[31m ♦ \033[0m",
		"\033[38;5;214m ▲ \033[0m",
		"\033[33m ★ \033[0m",
		"\033[32m ■ \033[0m",
		"\033[34m ▼ \033[0m",
		"\033[35m ● \033[0m",
		"\033[37m ✦ \033[0m",
	}
	if int(g) < len(ascii) {
		return ascii[g]
	}
	return " ? "
}

func gemToPlainText(g GemColor) string {
	chars := []string{" . ", " R ", " O ", " Y ", " G ", " B ", " P ", " W "}
	if int(g) < len(chars) {
		return chars[g]
	}
	return " ? "
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

// clearMatches finds active 3+ streaks and clears those gems AND everything above them.
// This acts as a lightweight gravity simulator so subsequent queued moves ignore falling zones.
func clearMatches(grid *[8][8]GemColor) bool {
	var toClear [8][8]bool
	foundAny := false

	for r := range 8 {
		for c := range 6 {
			color := grid[r][c]
			if color == Empty {
				continue
			}
			if grid[r][c+1] == color && grid[r][c+2] == color {
				toClear[r][c] = true
				toClear[r][c+1] = true
				toClear[r][c+2] = true
				foundAny = true
				for i := c + 3; i < 8 && grid[r][i] == color; i++ {
					toClear[r][i] = true
				}
			}
		}
	}

	for c := range 8 {
		for r := range 6 {
			color := grid[r][c]
			if color == Empty {
				continue
			}
			if grid[r+1][c] == color && grid[r+2][c] == color {
				toClear[r][c] = true
				toClear[r+1][c] = true
				toClear[r+2][c] = true
				foundAny = true
				for i := r + 3; i < 8 && grid[i][c] == color; i++ {
					toClear[i][c] = true
				}
			}
		}
	}

	if !foundAny {
		return false
	}

	for r := range 8 {
		for c := range 8 {
			if toClear[r][c] {
				for i := r; i >= 0; i-- {
					grid[i][c] = Empty
				}
			}
		}
	}

	return true
}

func checkLineMatch(grid [8][8]GemColor, r, c int) (bool, int) {
	color := grid[r][c]
	if color == Empty {
		return false, 99
	}

	left := c
	for left > 0 && grid[r][left-1] == color {
		left--
	}
	right := c
	for right < 7 && grid[r][right+1] == color {
		right++
	}
	horiz := right - left + 1

	up := r
	for up > 0 && grid[up-1][c] == color {
		up--
	}
	down := r
	for down < 7 && grid[down+1][c] == color {
		down++
	}
	vert := down - up + 1

	is5 := horiz >= 5 || vert >= 5
	isLOrT := horiz >= 3 && vert >= 3
	is4 := horiz == 4 || vert == 4
	is3 := horiz >= 3 || vert >= 3

	if is5 {
		return true, 2
	}
	if isLOrT {
		return true, 3
	}
	if is4 {
		return true, 4
	}
	if is3 {
		return true, 5
	}

	return false, 99
}

func testSwap(gridCopy [8][8]GemColor, r1, c1, r2, c2 int) (bool, int) {
	gridCopy[r1][c1], gridCopy[r2][c2] = gridCopy[r2][c2], gridCopy[r1][c1]

	match1, p1 := checkLineMatch(gridCopy, r1, c1)
	match2, p2 := checkLineMatch(gridCopy, r2, c2)

	if match1 || match2 {
		if match1 && match2 {
			if p1 < p2 {
				return true, p1
			}
			return true, p2
		}
		if match1 {
			return true, p1
		}
		return true, p2
	}

	return false, 99
}

func applyMove(grid *[8][8]GemColor, r1, c1, r2, c2 int) {
	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]
	clearMatches(grid)
}

func findAllMoves(grid [8][8]GemColor) []Move {
	var moves []Move
	gridPtr := &grid

	// Wipe out mid-animation exploding gems from the board evaluation
	clearMatches(gridPtr)

	// Search for priorities 2 through 5
	for priority := 2; priority <= 5; priority++ {
		for r := range 8 {
			for c := range 8 {
				if gridPtr[r][c] == Empty {
					continue
				}

				// Check Right Swap
				if c < 7 && gridPtr[r][c+1] != Empty {
					if isValid, p := testSwap(*gridPtr, r, c, r, c+1); isValid && p == priority {
						applyMove(gridPtr, r, c, r, c+1)
						moves = append(moves, Move{Row: r, Col: c, Dir: "Right", Priority: p})
						continue // Move to next cell since this one just exploded
					}
				}

				// Check Down Swap
				if r < 7 && gridPtr[r+1][c] != Empty {
					if isValid, p := testSwap(*gridPtr, r, c, r+1, c); isValid && p == priority {
						applyMove(gridPtr, r, c, r+1, c)
						moves = append(moves, Move{Row: r, Col: c, Dir: "Down", Priority: p})
					}
				}
			}
		}
	}

	return moves
}
