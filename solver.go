package main

import (
	"cmp"
	"slices"
)

// GetSortedMoves evaluates all possible moves on the board and sorts them by score.
func GetSortedMoves(b *Board) []Move {
	legalMoves := b.GenerateLegalMoves()

	slices.SortFunc(legalMoves, func(a, b Move) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return legalMoves
}

// GenerateLegalMoves scans the board and returns a list of all valid moves.
func (b *Board) GenerateLegalMoves() []Move {
	var legalMoves []Move

	// Create a copy of the board state to simulate making moves safely
	simState := b.State
	// Mark matched gems to not interfere with them
	markMatchedGems(&simState)

	// Iterate through all 64 spaces on the 8x8 grid
	for y := range 8 {
		for x := range 8 {
			idx := y*8 + x

			// 1. Try swapping Right (if we are not on the right edge)
			if x < 7 {
				rightIdx := idx + 1
				score := b.evaluateSwap(idx, rightIdx, simState)
				if score > 0 {
					legalMoves = append(legalMoves, Move{
						Index1: idx,
						Index2: rightIdx,
						Score:  score,
					})
				}
			}

			// 2. Try swapping Down (if we are not on the bottom edge)
			if y < 7 {
				downIdx := idx + 8
				score := b.evaluateSwap(idx, downIdx, simState)
				if score > 0 {
					legalMoves = append(legalMoves, Move{
						Index1: idx,
						Index2: downIdx,
						Score:  score,
					})
				}
			}
		}
	}

	legalMoves = append(legalMoves, Move{})

	return legalMoves
}

func markMatchedGems(grid *[64]Gem) {
	for y := range 8 {
		for x := range 6 {
			gem := grid[y*8+x]
			if gem.Color == ColorNone {
				continue
			}
			if grid[y*8+x+1].Color == gem.Color && grid[y*8+x+2].Color == gem.Color {
				grid[y*8+x].IsMatched = true
				grid[y*8+x+1].IsMatched = true
				grid[y*8+x+2].IsMatched = true
				for i := x + 3; i < 8 && grid[y*8+i].Color == gem.Color; i++ {
					grid[y*8+i].IsMatched = true
				}
			}
		}
	}

	for x := range 8 {
		for y := range 6 {
			gem := grid[y*8+x]
			if gem.Color == ColorNone {
				continue
			}
			if grid[(y+1)*8+x].Color == gem.Color && grid[(y+2)*8+x].Color == gem.Color {
				grid[y*8+x].IsMatched = true
				grid[(y+1)*8+x].IsMatched = true
				grid[(y+2)*8+x].IsMatched = true
				for i := y + 3; i < 8 && grid[i*8+x].Color == gem.Color; i++ {
					grid[i*8+x].IsMatched = true
				}
			}
		}
	}
}

func checkLineMatch(grid *[64]Gem, idx int) (MatchType, []int) {
	gem := grid[idx]
	if gem.Color == ColorNone || gem.IsMatched {
		return MatchTypeNone, nil
	}

	x := idx % 8
	y := idx / 8

	left := x
	for left > 0 && grid[y*8+left-1].Color == gem.Color && !grid[y*8+left-1].IsMatched {
		left--
	}
	right := x
	for right < 7 && grid[y*8+right+1].Color == gem.Color && !grid[y*8+right+1].IsMatched {
		right++
	}
	horiz := right - left + 1

	up := y
	for up > 0 && grid[(up-1)*8+x].Color == gem.Color && !grid[(up-1)*8+x].IsMatched {
		up--
	}
	down := y
	for down < 7 && grid[(down+1)*8+x].Color == gem.Color && !grid[(down+1)*8+x].IsMatched {
		down++
	}
	vert := down - up + 1

	is5 := horiz >= 5 || vert >= 5
	isLOrT := horiz >= 3 && vert >= 3
	is4 := horiz == 4 || vert == 4
	is3 := horiz >= 3 || vert >= 3

	// If there's no valid match of 3 or more in either direction, exit early
	if !is3 {
		return MatchTypeNone, nil
	}

	// Collect matched indices
	var matchedIndices []int

	if horiz >= 3 {
		for i := left; i <= right; i++ {
			matchedIndices = append(matchedIndices, y*8+i)
		}
	}

	if vert >= 3 {
		for j := up; j <= down; j++ {
			// Skip the center node (idx) if it was already added by horizontal check
			if horiz >= 3 && j == y {
				continue
			}
			matchedIndices = append(matchedIndices, j*8+x)
		}
	}

	// Determine match type
	if is5 {
		return MatchType5, matchedIndices
	}
	if isLOrT {
		return MatchTypeL, matchedIndices
	}
	if is4 {
		return MatchType4, matchedIndices
	}

	return MatchType3, matchedIndices
}

func (b *Board) evaluateSwap(idx1, idx2 int, state [64]Gem) int {
	gem1 := state[idx1]
	gem2 := state[idx2]

	if gem1.Color == ColorNone || gem2.Color == ColorNone || gem1.IsMatched || gem2.IsMatched {
		return 0
	}

	if gem1.State == StateHypercube && gem2.State == StateHypercube {
		return 10000
	}

	var destroyedGemIdxs []int

	if gem1.State == StateHypercube {
		destroyedGemIdxs = append(destroyedGemIdxs, idx1)

		return evaluateCascade(destroyedGemIdxs, &state, []GemColor{gem2.Color})
	}

	if gem2.State == StateHypercube {
		destroyedGemIdxs = append(destroyedGemIdxs, idx2)

		return evaluateCascade(destroyedGemIdxs, &state, []GemColor{gem1.Color})
	}

	state[idx1], state[idx2] = state[idx2], state[idx1]

	// Check if swapping them creates a match at either of their new positions
	matchType1, matchIdxs1 := checkLineMatch(&state, idx1)
	matchType2, matchIdxs2 := checkLineMatch(&state, idx2)

	swapScore := max(0, int(matchType1)-1)*10 + max(0, int(matchType2)-1)*10
	for i := range matchIdxs1 {
		destroyedGemIdxs = append(destroyedGemIdxs, i)
	}

	for i := range matchIdxs2 {
		destroyedGemIdxs = append(destroyedGemIdxs, i)
	}

	var colorTriggers []GemColor

	if matchType1 != MatchTypeNone && b.IsBlazingSpeed {
		match1ExplosionIds := getExplosionIdxs(&state, idx1)
		destroyedGemIdxs = append(destroyedGemIdxs, match1ExplosionIds...)

		for i := range match1ExplosionIds {
			if state[i].State == StateHypercube {
				colorTriggers = append(colorTriggers, state[idx1].Color)
			}
		}
	}

	if matchType2 != MatchTypeNone && b.IsBlazingSpeed {
		match2ExplosionIds := getExplosionIdxs(&state, idx2)
		destroyedGemIdxs = append(destroyedGemIdxs, match2ExplosionIds...)

		for i := range match2ExplosionIds {
			if state[i].State == StateHypercube {
				colorTriggers = append(colorTriggers, state[idx2].Color)
			}
		}
	}

	if len(destroyedGemIdxs) == 0 {
		return 0
	}

	slices.Sort(destroyedGemIdxs)
	destroyedGemIdxs = slices.Compact(destroyedGemIdxs)

	slices.Sort(colorTriggers)
	colorTriggers = slices.Compact(colorTriggers)

	swapScore += evaluateCascade(destroyedGemIdxs, &state, colorTriggers)

	return int(swapScore)
}

func getExplosionIdxs(currentState *[64]Gem, explosionIdx int) []int {
	row := explosionIdx / 8
	col := explosionIdx % 8

	// Allocate capacity up to 8 neighbors
	neighbors := make([]int, 0, 8)

	// Check all 8 direction offsets
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			// Skip the center cell
			if dr == 0 && dc == 0 {
				continue
			}

			newRow := row + dr
			newCol := col + dc

			// Verify row and column boundaries
			if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
				newIndex := newRow*8 + newCol
				// Verify the gem exists at index
				if !currentState[newIndex].IsEmpty() {
					neighbors = append(neighbors, newIndex)
				}
			}
		}
	}

	return neighbors
}

func getStarIdxs(currentState *[64]Gem, starIdx int) []int {
	row := starIdx / 8
	col := starIdx % 8

	// Max 14 possible line-of-sight neighbors in an 8x8 grid (7 horizontal + 7 vertical)
	rays := make([]int, 0, 14)

	// Up
	for r := row - 1; r >= 0; r-- {
		newIdx := r*8 + col
		// Verify the gem exists at index, !Gem.IsEmpty() helper?
		if !currentState[newIdx].IsEmpty() {
			rays = append(rays, newIdx)
		}
	}

	// Down
	for r := row + 1; r < 8; r++ {
		newIdx := r*8 + col
		// Verify the gem exists at index, !Gem.IsEmpty() helper?
		if !currentState[newIdx].IsEmpty() {
			rays = append(rays, newIdx)
		}
	}

	// Left
	for c := col - 1; c >= 0; c-- {
		newIdx := row*8 + c
		// Verify the gem exists at index, !Gem.IsEmpty() helper?
		if !currentState[newIdx].IsEmpty() {
			rays = append(rays, newIdx)
		}
	}

	// Right
	for c := col + 1; c < 8; c++ {
		newIdx := row*8 + c
		// Verify the gem exists at index, !Gem.IsEmpty() helper?
		if !currentState[newIdx].IsEmpty() {
			rays = append(rays, newIdx)
		}
	}

	return rays
}

func (g *Gem) Clear() {
	g.Color = ColorNone
	g.State = StateNormal
}

func (g *Gem) IsEmpty() bool {
	return g.Color == ColorNone && g.State == StateNormal || g.IsMatched
}

func evaluateCascade(destroyedGemIdxs []int, currentState *[64]Gem, colorTriggers []GemColor) int {
	score := 0

	for i := range destroyedGemIdxs {
		if currentState[i].IsEmpty() {
			continue
		}

		score++
		score += int(currentState[i].BonusTimeAmnt) * 1000

		newTriggerColor := currentState[i].Color

		newDestroyedGemIdxs := []int{}
		switch currentState[i].State {
		case StateFire:
			newDestroyedGemIdxs = getExplosionIdxs(currentState, i)
		case StateHypercube:
			for j := range 64 {
				g := currentState[j]

				if slices.Contains(colorTriggers, g.Color) {
					newDestroyedGemIdxs = append(newDestroyedGemIdxs, j)
				}
			}
		case StateStar:
			newDestroyedGemIdxs = getStarIdxs(currentState, i)
		}

		currentState[i].Clear()

		score += evaluateCascade(newDestroyedGemIdxs, currentState, []GemColor{newTriggerColor})
	}

	return score
}
