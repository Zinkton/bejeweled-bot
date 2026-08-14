package main

import (
	"fmt"
	"strings"
)

// ANSI color codes for terminal display
const (
	ansiReset     = "\033[0m"
	ansiDim       = "\033[2m"
	ansiBold      = "\033[1m"
	ansiRed       = "\033[1;31m"
	ansiWhite     = "\033[1;37m"
	ansiGreen     = "\033[1;32m"
	ansiYellow    = "\033[1;33m"
	ansiPurple    = "\033[1;35m"
	ansiOrange    = "\033[38;5;208m" // 256-color orange
	ansiBlue      = "\033[1;34m"
	ansiHypercube = "\033[1;36m" // Cyan / Prismatic
)

// PrintState renders an 8x8 ASCII/Unicode board with ANSI colors and a legend.
func PrintState(state *[64]Gem) {
	fmt.Println(FormatState(state, true))
}

// FormatState returns the board as a string. Set useColor to true for terminal colors.
func FormatState(state *[64]Gem, useColor bool) string {
	var sb strings.Builder

	// Top column coordinate header
	sb.WriteString("     0     1     2     3     4     5     6     7  \n")
	sb.WriteString("  ┌─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────┐\n")

	for y := 0; y < 8; y++ {
		// Row content
		sb.WriteString(fmt.Sprintf("%d │", y))
		for x := 0; x < 8; x++ {
			gem := state[y*8+x]
			cell := renderGemCell(gem, useColor)
			sb.WriteString(fmt.Sprintf("%s│", cell))
		}
		sb.WriteString("\n")

		// Row divider
		if y < 7 {
			sb.WriteString("  ├─────┼─────┼─────┼─────┼─────┼─────┼─────┼─────┤\n")
		} else {
			sb.WriteString("  └─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┘\n")
		}
	}

	// Legend
	sb.WriteString("\n" + ansiDim + "Legend:" + ansiReset + "\n")
	sb.WriteString("  Normal:  R   G   B   Y   P   O   W\n")
	sb.WriteString("  Special: [R] Fire  |  *R* Star  |  !R! Supernova  |  <H> Hypercube\n")
	sb.WriteString("  Bonus:   +N (e.g., R+5)\n")

	return sb.String()
}

// renderGemCell renders a 5-character cell string (e.g. "  R  ", " [R] ", " *G* ", " R+5 ")
func renderGemCell(g Gem, useColor bool) string {
	if g.IsEmpty() || g.Color == ColorNone {
		if useColor {
			return fmt.Sprintf("  %s·%s  ", ansiDim, ansiReset)
		}
		return "  ·  "
	}

	// 1. Color letter
	char := colorChar(g.Color)

	// 2. Format cell text based on State and Bonus Time
	var rawCell string
	switch g.State {
	case StateHypercube:
		rawCell = "<H>"
	case StateFire:
		rawCell = fmt.Sprintf("[%s]", char)
	case StateStar:
		rawCell = fmt.Sprintf("*%s*", char)
	case StateSupernova:
		rawCell = fmt.Sprintf("!%s!", char)
	default:
		if g.BonusTimeAmnt > 0 {
			rawCell = fmt.Sprintf("%s+%d", char, g.BonusTimeAmnt)
		} else {
			rawCell = fmt.Sprintf(" %s ", char)
		}
	}

	// Pad to 5 characters wide (centered)
	padded := centerPad(rawCell, 5)

	if !useColor {
		return padded
	}

	// 3. Apply ANSI terminal styling
	colorCode := colorToANSI(g.Color)
	if g.State == StateHypercube {
		colorCode = ansiHypercube
	}

	return fmt.Sprintf("%s%s%s", colorCode, padded, ansiReset)
}

func colorChar(c GemColor) string {
	switch c {
	case ColorRed:
		return "R"
	case ColorWhite:
		return "W"
	case ColorGreen:
		return "G"
	case ColorYellow:
		return "Y"
	case ColorPurple:
		return "P"
	case ColorOrange:
		return "O"
	case ColorBlue:
		return "B"
	default:
		return "?"
	}
}

func colorToANSI(c GemColor) string {
	switch c {
	case ColorRed:
		return ansiRed
	case ColorWhite:
		return ansiWhite
	case ColorGreen:
		return ansiGreen
	case ColorYellow:
		return ansiYellow
	case ColorPurple:
		return ansiPurple
	case ColorOrange:
		return ansiOrange
	case ColorBlue:
		return ansiBlue
	default:
		return ansiReset
	}
}

func centerPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	left := (width - len(s)) / 2
	right := width - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
