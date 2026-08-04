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
	Row      int
	Col      int
	Dir      string
	Priority int
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
	case Hypercube:
		return "\033[30m ■ \033[0m"
	default:
		return " ? "
	}
}

func gemToPlainText(g GemColor) string {
	switch g {
	case Empty:
		return " . "
	case Red:
		return " R "
	case Orange:
		return " O "
	case Yellow:
		return " Y "
	case Green:
		return " G "
	case Blue:
		return " B "
	case Purple:
		return " P "
	case White:
		return " W "
	case Hypercube:
		return " H "
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

func clearMatches(grid *[8][8]GemColor) bool {
	var toClear [8][8]bool
	foundAny := false

	for r := range 8 {
		for c := range 6 {
			color := grid[r][c]
			if color == Empty || color == Hypercube {
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
			if color == Empty || color == Hypercube {
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

func checkSwap(grid *[8][8]GemColor, r1, c1, r2, c2 int) bool {
	gem1 := grid[r1][c1]
	gem2 := grid[r2][c2]

	isHypercubeMove := false
	var targetColor GemColor = Empty

	if gem1 == Hypercube && gem2 != Empty {
		isHypercubeMove = true
		targetColor = gem2
	} else if gem2 == Hypercube && gem1 != Empty {
		isHypercubeMove = true
		targetColor = gem1
	}

	if isHypercubeMove {
		if gem1 == Hypercube && gem2 == Hypercube {
			for r := range 8 {
				for c := range 8 {
					grid[r][c] = Empty
				}
			}
		} else {
			grid[r1][c1] = Empty
			grid[r2][c2] = Empty

			for r := range 8 {
				for c := range 8 {
					if grid[r][c] == targetColor {
						grid[r][c] = Empty
					}
				}
			}
		}
		return true
	}

	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]

	if clearMatches(grid) {
		return true
	}

	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]
	return false
}

func evaluateCellMatch(grid *[8][8]GemColor, r, c int) int {
	if grid[r][c] == Empty || grid[r][c] == Hypercube {
		return 99
	}

	color := grid[r][c]

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
		return 2
	}
	if isLOrT {
		return 3
	}
	if is4 {
		return 4
	}
	if is3 {
		return 5
	}

	return 99
}

func calculatePriority(grid [8][8]GemColor, r1, c1, r2, c2 int) int {
	gem1 := grid[r1][c1]
	gem2 := grid[r2][c2]

	if gem1 == Hypercube && gem2 == Hypercube {
		return 1
	}

	if gem1 == Hypercube || gem2 == Hypercube {
		return 2
	}

	grid[r1][c1], grid[r2][c2] = grid[r2][c2], grid[r1][c1]

	p1 := evaluateCellMatch(&grid, r1, c1)
	p2 := evaluateCellMatch(&grid, r2, c2)

	if p1 < p2 {
		return p1
	}
	return p2
}

func findAllMoves(grid [8][8]GemColor) []Move {
	var moves []Move
	gridPtr := &grid

	clearMatches(gridPtr)

	for priority := 1; priority <= 5; priority++ {
		for r := range 8 {
			for c := range 8 {
				if gridPtr[r][c] == Empty {
					continue
				}

				if c < 7 && gridPtr[r][c+1] != Empty {
					gridCopy := *gridPtr
					if checkSwap(&gridCopy, r, c, r, c+1) {
						prio := calculatePriority(*gridPtr, r, c, r, c+1)
						if prio == priority {
							checkSwap(gridPtr, r, c, r, c+1)
							moves = append(moves, Move{Row: r, Col: c, Dir: "Right", Priority: prio})
						}
					}
				}

				if r < 7 && gridPtr[r+1][c] != Empty {
					gridCopy := *gridPtr
					if checkSwap(&gridCopy, r, c, r+1, c) {
						prio := calculatePriority(*gridPtr, r, c, r+1, c)
						if prio == priority {
							checkSwap(gridPtr, r, c, r+1, c)
							moves = append(moves, Move{Row: r, Col: c, Dir: "Down", Priority: prio})
						}
					}
				}
			}
		}
	}

	return moves
}
