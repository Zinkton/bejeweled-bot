package main

import (
	"testing"
)

// Helper to quickly build a board for testing using just color integers.
// -1 means empty/none.
func buildTestBoard(layout [64]int, states [64]GemState, bonusTimes [64]uint32) Board {
	var b Board
	for i, colorVal := range layout {
		if colorVal == -1 {
			b.State[i] = Gem{Color: ColorNone}
		} else {
			b.State[i] = Gem{Color: GemColor(colorVal), State: states[i], BonusTimeAmnt: bonusTimes[i]}
		}
	}
	return b
}

func testFindBestMove(t *testing.T, layout [64]int, states [64]GemState, bonusTimes [64]uint32, x1, y1, x2, y2 int) {
	board := buildTestBoard(layout, states, bonusTimes)

	bestMove := GetSortedMoves(&board)[0]

	expectedIdx1 := (y1 * 8) + x1
	expectedIdx2 := (y2 * 8) + x2

	actualX1, actualY1 := bestMove.Index1%8, bestMove.Index1/8
	actualX2, actualY2 := bestMove.Index2%8, bestMove.Index2/8

	if (bestMove.Index1 != expectedIdx1 || bestMove.Index2 != expectedIdx2) &&
		(bestMove.Index1 != expectedIdx2 || bestMove.Index2 != expectedIdx1) {
		t.Errorf(
			"Expected move between (%d,%d) and (%d,%d) [Indices %d, %d], got (%d,%d) and (%d,%d) [Indices %d, %d]",
			x1, y1, x2, y2, expectedIdx1, expectedIdx2,
			actualX1, actualY1, actualX2, actualY2, bestMove.Index1, bestMove.Index2,
		)
	}
}

func TestFindBestMove3Match(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 0, 0, 1, 0)
}

func TestFindBestMove4Match(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 2, 2, 3, 3, 4, 1,
	}

	states := [64]GemState{}
	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 1, 6, 1, 7)
}

func TestFindBestMoveLMatch(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 2, 4,
		2, 3, 4, 1, 2, 3, 3, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 3, 3, 4, 3,
		1, 2, 3, 4, 1, 2, 2, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 2, 2, 3, 3, 4, 1,
	}

	states := [64]GemState{}
	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 7, 3, 6, 3)
}

func TestFindBestMove5Match(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 2, 4,
		2, 3, 4, 1, 2, 3, 3, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		1, 3, 4, 1, 3, 3, 4, 3,
		2, 1, 3, 4, 1, 2, 2, 4,
		1, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 2, 2, 3, 3, 4, 1,
	}

	states := [64]GemState{}
	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 1, 4, 0, 4)
}

func TestFindBestMoveFireGem(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire

	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 0, 7, 1, 7)
}

func TestFindBestMoveStarGem(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+7] = StateStar

	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 6, 5, 7, 5)
}

func TestFindBestMoveHyperCube(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 3, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 0,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 3,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+5] = StateStar
	states[1*8+6] = StateHypercube

	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 6, 0, 6, 1)
}

func TestFindBestMoveDoubleHyperCube(t *testing.T) {
	layout := [64]int{
		1, 0, 1, 1, 3, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 0, 0,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+7] = StateStar
	states[1*8+6] = StateHypercube
	states[7*8+6] = StateHypercube
	states[7*8+7] = StateHypercube

	bonusTimes := [64]uint32{}
	bonusTimes[0] = 10

	testFindBestMove(t, layout, states, bonusTimes, 6, 7, 7, 7)
}

func TestFindBestMovePlus5(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+5] = StateStar
	states[1*8+6] = StateHypercube

	bonusTimes := [64]uint32{}
	bonusTimes[2] = 5

	testFindBestMove(t, layout, states, bonusTimes, 0, 0, 1, 0)
}

func TestFindBestMovePlus10(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+5] = StateStar
	states[1*8+6] = StateHypercube

	bonusTimes := [64]uint32{}
	bonusTimes[2] = 5
	bonusTimes[7*8+3] = 10

	testFindBestMove(t, layout, states, bonusTimes, 0, 7, 1, 7)
}

func TestFindBestMovePlus10Deep(t *testing.T) {
	layout := [64]int{
		0, 1, 0, 0, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		1, 0, 0, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 1,
		0, 2, 3, 4, 1, 2, 3, 4,
		2, 3, 4, 1, 0, 0, 4, 0,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 1,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+5] = StateStar
	states[1*8+6] = StateHypercube

	bonusTimes := [64]uint32{}
	bonusTimes[2] = 5
	bonusTimes[2*8+1] = 10

	testFindBestMove(t, layout, states, bonusTimes, 0, 7, 1, 7)
}

func TestFindBestMove6MatchDeep(t *testing.T) {
	layout := [64]int{
		2, 1, 2, 3, 1, 2, 3, 4,
		2, 3, 4, 1, 2, 3, 4, 3,
		3, 6, 5, 5, 1, 2, 0, 4,
		5, 5, 0, 0, 5, 5, 4, 1,
		1, 2, 3, 4, 1, 2, 3, 0,
		2, 3, 4, 1, 6, 6, 4, 6,
		1, 2, 3, 4, 1, 2, 3, 4,
		0, 1, 0, 0, 2, 3, 4, 3,
	}

	states := [64]GemState{}
	states[7*8+2] = StateFire
	states[5*8+5] = StateStar
	states[1*8+6] = StateHypercube

	bonusTimes := [64]uint32{}

	testFindBestMove(t, layout, states, bonusTimes, 6, 1, 6, 2)
}
