package main

import (
	"cmp"
	"slices"
)

// GetSortedMoves evaluates all possible moves on the board and sorts them by score.
func GetSortedMoves(b *Board, depth int) []Move {
	// Mark matched gems to not interfere with them
	clearMatchedGems(&b.State)

	legalMoves := b.GenerateLegalMoves()

	// Deep search
	for i, legalMove := range legalMoves {
		legalMoves[i].Score = (depth+1)*b.Push(legalMove) + getBestMoveScore(b, depth)
		b.Pop()
	}

	slices.SortFunc(legalMoves, func(a, b Move) int {
		return cmp.Compare(b.Score, a.Score)
	})

	return legalMoves
}

func getBestMoveScore(b *Board, depthLeft int) int {
	// No change after last move, leaf
	if b.StateHistory[len(b.StateHistory)-1] == b.State || depthLeft == 0 {
		return 0
	}

	legalMoves := b.GenerateLegalMoves()

	bestScore := -100
	for _, legalMove := range legalMoves {
		score := depthLeft*b.Push(legalMove) + getBestMoveScore(b, depthLeft-1)
		if score > bestScore {
			bestScore = score
		}
		b.Pop()
	}

	return bestScore
}

func (b *Board) Push(move Move) int {
	b.StateHistory = append(b.StateHistory, b.State)

	// Empty move, just settle
	if move.Index1 == move.Index2 {
		return b.Settle() - 5
	}

	return b.evaluateSwap(move.Index1, move.Index2, &b.State)
}

func (b *Board) Pop() {
	b.State = b.StateHistory[len(b.StateHistory)-1]
	b.StateHistory = b.StateHistory[:len(b.StateHistory)-1]
}

type matchSpan struct {
	isHoriz bool
	fixed   int // row index for horiz, col index for vert
	start   int // start index (col or row)
	end     int // end index (col or row)
	color   GemColor
	length  int
	used    bool
}

type settleEvent struct {
	centerIdx int
	matchType MatchType
	color     GemColor
	special   GemState // StateNormal if 3-match
}

// applyGravity pulls gems downward to fill empty slots beneath them in O(64) time.
func applyGravity(state *[64]Gem) {
	for x := range 8 {
		writeY := 7
		for readY := 7; readY >= 0; readY-- {
			readIdx := readY*8 + x
			if !state[readIdx].IsEmpty() {
				if writeY != readY {
					writeIdx := writeY*8 + x
					state[writeIdx] = state[readIdx]
					state[readIdx].Clear()
				}
				writeY--
			}
		}
	}
}

func (b *Board) findHorizontalSpans() []matchSpan {
	var spans []matchSpan
	for y := range 8 {
		x := 0
		for x < 8 {
			idx := y*8 + x
			if b.State[idx].IsEmpty() {
				x++
				continue
			}
			color := b.State[idx].Color
			end := x
			for end+1 < 8 && b.State[y*8+end+1].Color == color {
				end++
			}
			length := end - x + 1
			if length >= 3 {
				spans = append(spans, matchSpan{
					isHoriz: true,
					fixed:   y,
					start:   x,
					end:     end,
					color:   color,
					length:  length,
				})
			}
			x = end + 1
		}
	}
	return spans
}

func (b *Board) findVerticalSpans() []matchSpan {
	var spans []matchSpan
	for x := range 8 {
		y := 0
		for y < 8 {
			idx := y*8 + x
			if b.State[idx].IsEmpty() {
				y++
				continue
			}
			color := b.State[idx].Color
			end := y
			for end+1 < 8 && b.State[(end+1)*8+x].Color == color {
				end++
			}
			length := end - y + 1
			if length >= 3 {
				spans = append(spans, matchSpan{
					isHoriz: false,
					fixed:   x,
					start:   y,
					end:     end,
					color:   color,
					length:  length,
				})
			}
			y = end + 1
		}
	}
	return spans
}

func (b *Board) Settle() int {
	totalScore := 0

	for {
		// 1. O(N) In-place Gravity
		applyGravity(&b.State)

		// 2. Scan board spans
		horizSpans := b.findHorizontalSpans()
		vertSpans := b.findVerticalSpans()

		if len(horizSpans) == 0 && len(vertSpans) == 0 {
			break
		}

		var destroyedGemIdxs []int
		var events []settleEvent
		var matchScore int

		// 3. Find L/T Intersections (Star Gems)
		for i := range horizSpans {
			h := &horizSpans[i]
			for j := range vertSpans {
				v := &vertSpans[j]
				if h.color == v.color && v.fixed >= h.start && v.fixed <= h.end && h.fixed >= v.start && h.fixed <= v.end {
					h.used = true
					v.used = true
					intersectIdx := h.fixed*8 + v.fixed

					events = append(events, settleEvent{
						centerIdx: intersectIdx,
						matchType: MatchTypeL,
						color:     h.color,
						special:   StateStar,
					})
				}
			}
		}

		// 4. Horizontal Spans
		for _, h := range horizSpans {
			for col := h.start; col <= h.end; col++ {
				destroyedGemIdxs = append(destroyedGemIdxs, h.fixed*8+col)
			}
			if h.used {
				continue
			}

			var special GemState = StateNormal
			var mType MatchType = MatchType3
			centerCol := h.start + h.length/2

			switch {
			case h.length >= 6:
				special, mType = StateSupernova, MatchType6
			case h.length == 5:
				special, mType = StateHypercube, MatchType5
			case h.length == 4:
				special, mType = StateFire, MatchType4
				centerCol = h.start + 1
			}

			events = append(events, settleEvent{
				centerIdx: h.fixed*8 + centerCol,
				matchType: mType,
				color:     h.color,
				special:   special,
			})
		}

		// 5. Vertical Spans
		for _, v := range vertSpans {
			for row := v.start; row <= v.end; row++ {
				destroyedGemIdxs = append(destroyedGemIdxs, row*8+v.fixed)
			}
			if v.used {
				continue
			}

			var special GemState = StateNormal
			var mType MatchType = MatchType3
			centerRow := v.start + v.length/2

			switch {
			case v.length >= 6:
				special, mType = StateSupernova, MatchType6
			case v.length == 5:
				special, mType = StateHypercube, MatchType5
			case v.length == 4:
				special, mType = StateFire, MatchType4
				centerRow = v.start + 1
			}

			events = append(events, settleEvent{
				centerIdx: centerRow*8 + v.fixed,
				matchType: mType,
				color:     v.color,
				special:   special,
			})
		}

		// Accumulate match scores
		for _, ev := range events {
			matchScore += max(0, int(ev.matchType)-1) * 10
		}

		// 6. Blazing Speed explosions (evaluated for all match centers)
		var colorTriggers []GemColor
		if b.IsBlazingSpeed {
			for _, ev := range events {
				expIds := getExplosionIdxs(&b.State, ev.centerIdx)
				destroyedGemIdxs = append(destroyedGemIdxs, expIds...)
				for _, expIdx := range expIds {
					if b.State[expIdx].State == StateHypercube {
						colorTriggers = append(colorTriggers, ev.color)
					}
				}
			}
		}

		// Deduplicate
		slices.Sort(destroyedGemIdxs)
		destroyedGemIdxs = slices.Compact(destroyedGemIdxs)

		if len(colorTriggers) > 0 {
			slices.Sort(colorTriggers)
			colorTriggers = slices.Compact(colorTriggers)
		}

		// 7. Cascade resolution
		cascadeScore := evaluateCascade(destroyedGemIdxs, &b.State, colorTriggers)
		totalScore += matchScore + cascadeScore

		// 8. Spawn special gems
		for _, ev := range events {
			if ev.special != StateNormal {
				b.State[ev.centerIdx] = Gem{Color: ev.color, State: ev.special}
			}
		}
	}

	return totalScore
}

// GenerateLegalMoves scans the board and returns a list of all valid moves.
func (b *Board) GenerateLegalMoves() []Move {
	var legalMoves []Move

	// Iterate through all 64 spaces on the 8x8 grid
	for y := range 8 {
		for x := range 8 {
			idx := y*8 + x

			// 1. Try swapping Right (if we are not on the right edge)
			if x < 7 {
				rightIdx := idx + 1
				if b.isLegalMove(idx, rightIdx, b.State) {
					legalMoves = append(legalMoves, Move{
						Index1: idx,
						Index2: rightIdx,
					})
				}
			}

			// 2. Try swapping Down (if we are not on the bottom edge)
			if y < 7 {
				downIdx := idx + 8
				if b.isLegalMove(idx, downIdx, b.State) {
					legalMoves = append(legalMoves, Move{
						Index1: idx,
						Index2: downIdx,
					})
				}
			}
		}
	}

	// Prefer not waiting
	legalMoves = append(legalMoves, Move{Score: -5})

	return legalMoves
}

func clearMatchedGems(grid *[64]Gem) {
	for y := range 8 {
		for x := range 6 {
			gem := grid[y*8+x]
			if gem.IsEmpty() {
				continue
			}
			if grid[y*8+x+1].Color == gem.Color && grid[y*8+x+2].Color == gem.Color {
				grid[y*8+x].ClearNormal()
				grid[y*8+x+1].ClearNormal()
				grid[y*8+x+2].ClearNormal()
				for i := x + 3; i < 8 && grid[y*8+i].Color == gem.Color; i++ {
					grid[y*8+i].ClearNormal()
				}
			}
		}
	}

	for x := range 8 {
		for y := range 6 {
			gem := grid[y*8+x]
			if gem.IsEmpty() {
				continue
			}
			if grid[(y+1)*8+x].Color == gem.Color && grid[(y+2)*8+x].Color == gem.Color {
				grid[y*8+x].ClearNormal()
				grid[(y+1)*8+x].ClearNormal()
				grid[(y+2)*8+x].ClearNormal()
				for i := y + 3; i < 8 && grid[i*8+x].Color == gem.Color; i++ {
					grid[i*8+x].ClearNormal()
				}
			}
		}
	}
}

func checkLineMatch(grid *[64]Gem, idx int) (MatchType, []int) {
	gem := grid[idx]
	if gem.IsEmpty() {
		return MatchTypeNone, nil
	}

	x := idx % 8
	y := idx / 8

	left := x
	for left > 0 && grid[y*8+left-1].Color == gem.Color {
		left--
	}
	right := x
	for right < 7 && grid[y*8+right+1].Color == gem.Color {
		right++
	}
	horiz := right - left + 1

	up := y
	for up > 0 && grid[(up-1)*8+x].Color == gem.Color {
		up--
	}
	down := y
	for down < 7 && grid[(down+1)*8+x].Color == gem.Color {
		down++
	}
	vert := down - up + 1

	is6 := horiz >= 6 || vert >= 6
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
	if is6 {
		return MatchType6, matchedIndices
	}
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

// Evaluate and execute the swap
func (b *Board) isLegalMove(idx1, idx2 int, state [64]Gem) bool {
	gem1 := state[idx1]
	gem2 := state[idx2]

	if gem1.IsEmpty() || gem2.IsEmpty() {
		return false
	}

	if gem1.State == StateHypercube || gem2.State == StateHypercube {
		return true
	}

	state[idx1], state[idx2] = state[idx2], state[idx1]

	// Check if swapping them creates a match at either of their new positions
	matchType1, _ := checkLineMatch(&state, idx1)
	if matchType1 != MatchTypeNone {
		return true
	}

	matchType2, _ := checkLineMatch(&state, idx2)
	if matchType2 != MatchTypeNone {
		return true
	}

	return false
}

// Evaluate and execute the swap
func (b *Board) evaluateSwap(idx1, idx2 int, state *[64]Gem) int {
	gem1 := state[idx1]
	gem2 := state[idx2]

	if gem1.IsEmpty() || gem2.IsEmpty() {
		return 0
	}

	if gem1.State == StateHypercube && gem2.State == StateHypercube {
		return 100000
	}

	var destroyedGemIdxs []int

	if gem1.State == StateHypercube {
		destroyedGemIdxs = append(destroyedGemIdxs, idx1)
		score := evaluateCascade(destroyedGemIdxs, state, []GemColor{gem2.Color})
		applyGravity(state)

		return score
	}

	if gem2.State == StateHypercube {
		destroyedGemIdxs = append(destroyedGemIdxs, idx2)
		score := evaluateCascade(destroyedGemIdxs, state, []GemColor{gem1.Color})
		applyGravity(state)

		return score
	}

	specialCountBefore := 0
	for i := range state {
		if state[i].State != StateNormal {
			specialCountBefore++
		}
	}

	state[idx1], state[idx2] = state[idx2], state[idx1]

	// Check if swapping them creates a match at either of their new positions
	matchType1, matchIdxs1 := checkLineMatch(state, idx1)
	matchType2, matchIdxs2 := checkLineMatch(state, idx2)

	match1Color := state[idx1].Color
	match2Color := state[idx2].Color

	swapScore := max(0, int(matchType1)-1)*10 + max(0, int(matchType2)-1)*10
	for _, match1Idx := range matchIdxs1 {
		destroyedGemIdxs = append(destroyedGemIdxs, match1Idx)
	}

	for _, match2Idx := range matchIdxs2 {
		destroyedGemIdxs = append(destroyedGemIdxs, match2Idx)
	}

	var colorTriggers []GemColor
	isSpecialGemActivated := false
	if matchType1 != MatchTypeNone && b.IsBlazingSpeed {
		isSpecialGemActivated = true
		match1ExplosionIds := getExplosionIdxs(state, idx1)
		destroyedGemIdxs = append(destroyedGemIdxs, match1ExplosionIds...)

		for _, match1ExplosionIdx := range match1ExplosionIds {
			if state[match1ExplosionIdx].State == StateHypercube {
				colorTriggers = append(colorTriggers, state[idx1].Color)
			}
		}
	}

	if matchType2 != MatchTypeNone && b.IsBlazingSpeed {
		isSpecialGemActivated = true
		match2ExplosionIds := getExplosionIdxs(state, idx2)
		destroyedGemIdxs = append(destroyedGemIdxs, match2ExplosionIds...)

		for _, match2ExplosionIdx := range match2ExplosionIds {
			if state[match2ExplosionIdx].State == StateHypercube {
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

	swapScore += evaluateCascade(destroyedGemIdxs, state, colorTriggers)

	// After the cascade, the newly created special gems survive in bejeweled 3 I believe
	switch matchType1 {
	case MatchType4:
		state[idx1] = Gem{Color: match1Color, State: StateFire}
	case MatchTypeL:
		state[idx1] = Gem{Color: match1Color, State: StateStar}
	case MatchType5:
		state[idx1] = Gem{Color: match1Color, State: StateHypercube}
	case MatchType6:
		state[idx1] = Gem{Color: match1Color, State: StateSupernova}
	}

	switch matchType2 {
	case MatchType4:
		state[idx2] = Gem{Color: match2Color, State: StateFire}
	case MatchTypeL:
		state[idx2] = Gem{Color: match2Color, State: StateStar}
	case MatchType5:
		state[idx2] = Gem{Color: match2Color, State: StateHypercube}
	case MatchType6:
		state[idx2] = Gem{Color: match2Color, State: StateSupernova}
	}

	specialCountAfter := 0
	for i := range state {
		if state[i].State != StateNormal {
			specialCountAfter++
		}
	}

	if isSpecialGemActivated || specialCountAfter < specialCountBefore {
		applyGravity(state)
	}

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

func (g *Gem) ClearNormal() {
	if g.State == StateNormal {
		g.Clear()
	}
}

func (g *Gem) IsEmpty() bool {
	return g.Color == ColorNone && g.State == StateNormal
}

type CascadeEvent struct {
	Idx           int
	TriggerColors []GemColor
}

func evaluateCascade(initialIdxs []int, currentState *[64]Gem, initialTriggers []GemColor) int {
	score := 0
	var queue []CascadeEvent

	// 1. Seed the initial wave into the queue
	for _, idx := range initialIdxs {
		queue = append(queue, CascadeEvent{
			Idx:           idx,
			TriggerColors: initialTriggers,
		})
	}

	// 2. Process the queue until it's empty (Breadth-First)
	for len(queue) > 0 {
		// Pop the front event off the queue
		event := queue[0]
		queue = queue[1:]

		idx := event.Idx

		// If another explosion in this wave already cleared it, skip
		if currentState[idx].IsEmpty() {
			continue
		}

		// Score it
		score++
		score += int(currentState[idx].BonusTimeAmnt) * 1000

		// Save the color of THIS gem, because if this gem causes a
		// chain reaction, its color becomes the new trigger color.
		thisGemColor := currentState[idx].Color
		var nextIdxs []int

		// Determine blast radius
		switch currentState[idx].State {
		case StateFire:
			nextIdxs = getExplosionIdxs(currentState, idx)

		case StateHypercube:
			for j := range 64 {
				if slices.Contains(event.TriggerColors, currentState[j].Color) {
					nextIdxs = append(nextIdxs, j)
				}
			}

		case StateStar:
			nextIdxs = getStarIdxs(currentState, idx)

		case StateSupernova:
			explosionIdxs := getExplosionIdxs(currentState, idx)
			for _, expIdx := range explosionIdxs {
				nextIdxs = append(nextIdxs, getStarIdxs(currentState, expIdx)...)
			}
			slices.Sort(nextIdxs)
			nextIdxs = slices.Compact(nextIdxs)
		}

		// 3. Clear the gem IMMEDIATELY so it doesn't get re-queued by overlapping blasts
		currentState[idx].Clear()

		// 4. Queue up all the gems caught in the blast for the next wave
		for _, nextIdx := range nextIdxs {
			// Minor optimization: don't bother queuing empty spaces
			if !currentState[nextIdx].IsEmpty() {
				queue = append(queue, CascadeEvent{
					Idx:           nextIdx,
					TriggerColors: []GemColor{thisGemColor},
				})
			}
		}
	}

	return score
}
