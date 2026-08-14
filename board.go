package main

// --- Enums ---

type GemColor uint8

const (
	ColorRed    GemColor = 0
	ColorWhite  GemColor = 1
	ColorGreen  GemColor = 2
	ColorYellow GemColor = 3
	ColorPurple GemColor = 4
	ColorOrange GemColor = 5
	ColorBlue   GemColor = 6
	ColorNone   GemColor = 255 // Useful for empty/invalid spaces
)

type GemState uint8

const (
	StateNormal    GemState = 0
	StateFire      GemState = 1
	StateHypercube GemState = 2
	StateStar      GemState = 4
	StateSupernova GemState = 5
)

type MatchType uint8

const (
	MatchTypeNone MatchType = 0
	MatchType3    MatchType = 1
	MatchType4    MatchType = 2
	MatchTypeL    MatchType = 3
	MatchType5    MatchType = 4
	MatchType6    MatchType = 5
)

// --- Structs ---

type Gem struct {
	Color         GemColor
	State         GemState
	BonusTimeAmnt uint32
}

type Board struct {
	State          [64]Gem
	IsBlazingSpeed bool
	StateHistory   [][64]Gem
}

// Move represents a potential swap between two adjacent indices on the board
type Move struct {
	Index1 int
	Index2 int
	Score  int // The calculated value of making this move
}
