package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// --- Enums ---

type ViewMode int

const (
	ModeDebug ViewMode = iota
	ModeHintOverlay
)

type Game struct {
	Board              Board
	ProcessFound       bool
	IsActive           bool
	LastTimer          uint32
	CachedBestMove     *Move
	LastEvaluatedBoard [64]Gem
	HasCalculatedMove  bool
	Mode               ViewMode
	Geom               *WindowBounds
	BotEnabled         bool
	ScreenWidth        int
	ScreenHeight       int
}

// --- Main Loop ---

func (g *Game) Update() error {
	if !g.ProcessFound {
		if AttachToProcess("bejeweled3.exe") {
			g.ProcessFound = true
			fmt.Println("Process found! Base address:", fmt.Sprintf("%X", baseAddress))
		}
		return nil
	}

	g.syncGeometry()

	g.parseBoard()

	g.updateBestMove()

	return nil
}

// syncGeometry checks if Bejeweled has moved/resized and updates coordinates
func (g *Game) syncGeometry() {
	geom, err := GetGameGeometry()
	if err != nil {
		return
	}

	// Detect if position or dimensions changed
	if g.Geom == nil ||
		geom.ScreenX != g.Geom.ScreenX ||
		geom.ScreenY != g.Geom.ScreenY ||
		geom.ClientWidth != g.Geom.ClientWidth ||
		geom.ClientHeight != g.Geom.ClientHeight {

		g.Geom = geom

		// If overlay mode is active, re-snap Ebiten's window to match
		if g.Mode == ModeHintOverlay {
			snapEbitenWindow(geom.ScreenX, geom.ScreenY, geom.ClientWidth, geom.ClientHeight)
		}
	}
}

func (g *Game) drawDebugView(screen *ebiten.Image) {
	// 1. Draw Blazing Speed Tint (Background)
	if g.Board.IsBlazingSpeed {
		bounds := screen.Bounds()
		vector.FillRect(screen, 0, 0, float32(bounds.Dx()), float32(bounds.Dy()), color.NRGBA{255, 128, 0, 100}, false)
	}

	if !g.ProcessFound {
		ebitenutil.DebugPrint(screen, "Searching for bejeweled.exe...")
		return
	}

	ebitenutil.DebugPrint(screen, "Bejeweled 3 Board Data\n(Press Esc to exit)")

	// 2. Draw Paused state
	if !g.IsActive {
		ebitenutil.DebugPrintAt(screen, "=== PAUSED ===", 300, 20)
	}

	// 3. Draw the 8x8 grid
	startX, startY := 50, 50
	tileSize := float32(40)
	spacing := float32(5)

	for i, gem := range g.Board.State {
		x := float32(startX) + float32(i%8)*(tileSize+spacing)
		y := float32(startY) + float32(i/8)*(tileSize+spacing)

		// Draw the Gem Core (Shape or Hypercube)
		if gem.State == StateHypercube {
			drawHypercube(screen, x, y, tileSize)
		} else {
			gemColor := getGemColor(gem.Color)
			drawGemShape(screen, x, y, tileSize, gem.Color, gemColor)
		}

		// Draw State Outlines (Fire / Star / Supernova)
		switch gem.State {
		case StateFire:
			drawOutline(screen, x, y, tileSize, 5, color.RGBA{255, 0, 0, 255}) // Red outline
		case StateStar:
			drawOutline(screen, x, y, tileSize, 5, color.RGBA{0, 0, 255, 255}) // Blue outline
		case StateSupernova:
			drawOutline(screen, x, y, tileSize, 5, color.RGBA{255, 255, 255, 255}) // White outline
		}

		// Draw Bonus Time text
		if gem.BonusTimeAmnt > 0 {
			text := fmt.Sprintf("+%d", gem.BonusTimeAmnt)

			vector.FillRect(screen, x+4, y+10, 30, 16, color.NRGBA{0, 0, 0, 120}, false)
			ebitenutil.DebugPrintAt(screen, text, int(x)+6, int(y)+10)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.Mode == ModeHintOverlay {
		return outsideWidth, outsideHeight
	}

	return 640, 480
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.Mode {
	case ModeDebug:
		g.drawDebugView(screen)
	case ModeHintOverlay:
		g.drawHintOverlay(screen)
	}
}

func (g *Game) updateBestMove() {
	if g.HasCalculatedMove && g.Board.State == g.LastEvaluatedBoard || !g.IsActive {
		return
	}

	moves := GetSortedMoves(&g.Board, 2)

	if len(moves) > 0 {
		best := moves[0]
		g.CachedBestMove = &best
	} else {
		g.CachedBestMove = nil
	}

	g.LastEvaluatedBoard = g.Board.State
	g.HasCalculatedMove = true
}

// --- Memory parsing ---
func (g *Game) parseBoard() {
	if processHandle == 0 {
		return
	}

	// state = bejeweled.exe + 004E170C + 360 + A0
	ptr1, err := ReadUint32(baseAddress + 0x004E170C)
	if err != nil || ptr1 == 0 {
		return
	}

	ptr2, err := ReadUint32(ptr1 + 0x360)
	if err != nil || ptr2 == 0 {
		return
	}

	state, err := ReadUint32(ptr2 + 0xA0)
	if err != nil || state == 0 {
		return
	}

	// Blazing Speed (state + 10A0)
	blazing, _ := ReadUint32(state + 0x10A0)
	g.Board.IsBlazingSpeed = (blazing > 0)

	// Timer check (state + D68)
	timer, _ := ReadUint32(state + 0xE38)
	if timer != g.LastTimer {
		g.IsActive = true
		g.LastTimer = timer
	} else {
		g.IsActive = false
	}

	topLeftGemPtr := uint32(state + 0xF8)

	// Loop through all 64 gems (+4 bytes each)
	for i := range 64 {
		gemPtr, err := ReadUint32(topLeftGemPtr + uint32(i*4))
		if err != nil || gemPtr == 0 {
			// Clear it out if it fails to read (e.g. board is resetting)
			g.Board.State[i] = Gem{}
			continue
		}

		// Read these as 1 byte (uint8)
		colorVal, _ := ReadUint8(gemPtr + 0x220)
		stateVal, _ := ReadUint8(gemPtr + 0x228)
		otherStateVal, _ := ReadUint8(gemPtr + 0x22A)
		bonusTimeAmnt, _ := ReadUint8(gemPtr + 0x244)

		bonus := uint32(0)
		if otherStateVal == 2 {
			bonus = uint32(bonusTimeAmnt)
		}

		g.Board.State[i] = Gem{
			Color:         GemColor(colorVal),
			State:         GemState(stateVal),
			BonusTimeAmnt: bonus,
		}
	}
}

// --- Drawing Helpers ---

// drawOutline draws a hollow rectangle manually using 4 filled rectangles
func drawOutline(screen *ebiten.Image, x, y, size, thickness float32, clr color.Color) {
	vector.FillRect(screen, x, y, size, thickness, clr, false)                // Top
	vector.FillRect(screen, x, y+size-thickness, size, thickness, clr, false) // Bottom
	vector.FillRect(screen, x, y, thickness, size, clr, false)                // Left
	vector.FillRect(screen, x+size-thickness, y, thickness, size, clr, false) // Right
}

// drawHypercube draws 6 colored blocks smushed together, now with padding
func drawHypercube(screen *ebiten.Image, x, y, size float32) {
	p := float32(6) // Add padding so it doesn't take up the whole tile
	innerSize := size - p*2
	w := innerSize / 3
	h := innerSize / 2
	colors := []color.Color{
		getGemColor(ColorRed), getGemColor(ColorWhite), getGemColor(ColorGreen),
		getGemColor(ColorYellow), getGemColor(ColorPurple), getGemColor(ColorBlue),
	}

	// Draw the dark tile background first
	vector.FillRect(screen, x, y, size, size, color.RGBA{30, 30, 30, 255}, false)

	for i, clr := range colors {
		col := float32(i % 3)
		row := float32(i / 3)
		vector.FillRect(screen, x+p+col*w, y+p+row*h, w, h, clr, false)
	}
}

// drawGemShape uses simple geometry to differentiate gem colors visually
func drawGemShape(screen *ebiten.Image, x, y, size float32, c GemColor, clr color.Color) {
	p := float32(10)

	// Draw a dark background for the tile first
	vector.FillRect(screen, x, y, size, size, color.RGBA{30, 30, 30, 255}, false)

	switch c {
	case ColorRed:
		vector.FillRect(screen, x+p, y+p, size-p*2, size-p*2, clr, false)
	case ColorWhite:
		vector.FillRect(screen, x+size/2.5, y+size/2.5, size-(size/2.5)*2, size-(size/2.5)*2, clr, false)
	case ColorGreen:
		vector.FillRect(screen, x+size/2-p/2, y+p, p, size-p*2, clr, false)
	case ColorYellow:
		vector.FillRect(screen, x+p, y+size/2-p/2, size-p*2, p, clr, false)
	case ColorPurple:
		vector.FillRect(screen, x+size/2-p/4, y+p, p/2, size-p*2, clr, false)
		vector.FillRect(screen, x+p, y+size/2-p/4, size-p*2, p/2, clr, false)
	case ColorOrange:
		vector.FillRect(screen, x+p, y+p, p/1.5, size-p*2, clr, false)
		vector.FillRect(screen, x+size-p-p/1.5, y+p, p/1.5, size-p*2, clr, false)
	case ColorBlue:
		drawOutline(screen, x+p, y+p, size-p*2, 3, clr)
	default:
		vector.FillRect(screen, x+p, y+p, size-p*2, size-p*2, color.RGBA{100, 100, 100, 255}, false)
	}
}

func getGemColor(c GemColor) color.Color {
	switch c {
	case ColorRed:
		return color.RGBA{255, 50, 50, 255}
	case ColorWhite:
		return color.RGBA{255, 255, 255, 255}
	case ColorGreen:
		return color.RGBA{50, 255, 50, 255}
	case ColorYellow:
		return color.RGBA{255, 255, 50, 255}
	case ColorPurple:
		return color.RGBA{200, 50, 255, 255}
	case ColorOrange:
		return color.RGBA{255, 150, 50, 255}
	case ColorBlue:
		return color.RGBA{50, 150, 255, 255}
	default:
		return color.RGBA{50, 50, 50, 255}
	}
}

func main() {
	geom, err := GetGameGeometry()
	if err != nil {
		fmt.Println("[!] Could not auto-detect Bejeweled 3 window at launch. Will retry on toggle.")
	}

	game := &Game{
		Mode: ModeDebug,
		Geom: geom,
	}

	bot := &BotController{
		IsEnabled: false,
		Geom:      geom,
	}

	// Start bot execution worker & hotkeys
	game.StartBotWorker(bot)
	game.StartHotkeyListener(bot)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle(OverlayWindowTitle)

	// ScreenTransparent enables the transparent framebuffer for overlay mode
	options := &ebiten.RunGameOptions{
		ScreenTransparent: true,
	}

	if err := ebiten.RunGameWithOptions(game, options); err != nil {
		log.Fatal(err)
	}
}
