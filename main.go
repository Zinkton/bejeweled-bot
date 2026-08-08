package main

import (
	"fmt"
	"image/color"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-vgo/robotgo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.design/x/hotkey"
)

type Bot struct {
	WindowCloaked bool
	// Vision Bot Coordinates (Physical Pixels for robotgo)
	AnchorX, AnchorY int
	StepX, StepY     float64
	HasTopLeft       bool
	HasBotRight      bool

	// Visual Overlay Coordinates (Logical Pixels for Ebiten)
	VisualWidth  int
	VisualHeight int

	IsLocked     bool
	IsRunning    bool
	CurrentMoves []Move
	ActiveMoves  [2]*Move
}

func (b *Bot) Update() error {
	// The window has just spawned, apply the OS cloak!
	if !b.WindowCloaked {
		cloakWindow("BejeweledBotOverlay123")
		b.WindowCloaked = true
	}
	return nil
}

func cloakWindow(title string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	findWindow := user32.NewProc("FindWindowW")
	setAffinity := user32.NewProc("SetWindowDisplayAffinity")

	// 1. Find the window by its title
	ptr, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(ptr)))

	if hwnd != 0 {
		// 2. WDA_EXCLUDEFROMCAPTURE (0x00000011) hides it from screenshots
		setAffinity.Call(hwnd, 0x00000011)
		fmt.Println("[!] Window cloaked from screen capture APIs.")
	} else {
		fmt.Println("[!] Could not find window to cloak.")
	}
}

func (b *Bot) Draw(screen *ebiten.Image) {
	// --- SETUP MODE: Draw an alignment grid ---
	if !b.IsLocked {
		// ... (Keep your existing setup drawing code here) ...
		return
	}

	// --- RUN MODE: Draw the hint squares ---
	if !b.IsRunning {
		return
	}

	vStepX := float32(b.VisualWidth) / 8.0
	vStepY := float32(b.VisualHeight) / 8.0

	// Define our two colors: Green for the 1st move, Red for the 2nd
	colors := []color.NRGBA{
		{R: 0, G: 200, B: 0, A: 150},
		{R: 200, G: 0, B: 0, A: 150},
	}

	for i, activeMove := range b.ActiveMoves {
		// Skip if this slot is empty (e.g., only 1 possible move on the board)
		if activeMove == nil {
			continue
		}

		m := *activeMove
		c := colors[i] // Grab the color corresponding to this slot

		x := float32(m.Col) * vStepX
		y := float32(m.Row) * vStepY

		// Draw Source Gem
		vector.FillRect(screen, x, y, vStepX, vStepY, c, false)

		tr, tc := m.Row, m.Col
		switch m.Dir {
		case "Up":
			tr--
		case "Down":
			tr++
		case "Left":
			tc--
		case "Right":
			tc++
		}

		// Draw Target Gem
		if tr >= 0 && tr < 8 && tc >= 0 && tc < 8 {
			tx := float32(tc) * vStepX
			ty := float32(tr) * vStepY
			vector.FillRect(screen, tx, ty, vStepX, vStepY, c, false)
		}
	}
}

// Layout dynamically updates the visual width and height when you resize the window
func (b *Bot) Layout(outsideWidth, outsideHeight int) (int, int) {
	b.VisualWidth = outsideWidth
	b.VisualHeight = outsideHeight
	return outsideWidth, outsideHeight
}

func main() {
	bot := &Bot{}

	fmt.Println("=== Bejeweled Hybrid Overlay ===")
	fmt.Println("1. Drag the transparent window to your 2nd monitor.")
	fmt.Println("2. Resize it so the 8x8 grid perfectly covers the gems.")
	fmt.Println("3. Press 'Ctrl+Shift+L' to LOCK the window into an overlay.")

	// 1. Start setup window
	ebiten.SetWindowSize(600, 600)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) // Allow user to stretch it
	ebiten.SetWindowDecorated(true)                                // Give it a title bar so it can be dragged
	ebiten.SetWindowFloating(true)                                 // Always on top

	go bot.ListenForHotkeys()
	go bot.Worker()

	ebiten.SetWindowTitle("BejeweledBotOverlay123")

	options := &ebiten.RunGameOptions{ScreenTransparent: true}
	if err := ebiten.RunGameWithOptions(bot, options); err != nil {
		panic(err)
	}
}

func (b *Bot) ListenForHotkeys() {
	hkLock := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyL)
	hkX := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyX)
	hkC := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyC)
	hkZ := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyZ)

	hkLock.Register()
	defer hkLock.Unregister()
	hkX.Register()
	defer hkX.Unregister()
	hkC.Register()
	defer hkC.Unregister()
	hkZ.Register()
	defer hkZ.Unregister()

	for {
		select {
		case <-hkLock.Keydown():
			if !b.IsLocked {
				// Lock the window! Remove title bar and make it click-through
				ebiten.SetWindowDecorated(false)
				ebiten.SetWindowMousePassthrough(true)
				b.IsLocked = true
				fmt.Println("\n[✔] UI Locked!")
				fmt.Println("Now calibrate the vision bot:")
				fmt.Println("-> Hover TOP-LEFT gem and press 'Ctrl+Shift+X'")
			}

		case <-hkX.Keydown():
			if b.IsLocked {
				b.AnchorX, b.AnchorY = robotgo.Location()
				b.HasTopLeft = true
				fmt.Printf("[Calibrated] Top-Left at X:%d Y:%d\n", b.AnchorX, b.AnchorY)
				fmt.Println("-> Hover BOTTOM-RIGHT gem and press 'Ctrl+Shift+C'")
			}

		case <-hkC.Keydown():
			if b.HasTopLeft {
				x, y := robotgo.Location()
				b.StepX = float64(x-b.AnchorX) / 7.0
				b.StepY = float64(y-b.AnchorY) / 7.0
				b.HasBotRight = true
				fmt.Printf("[Calibrated] Bottom-Right at X:%d Y:%d\n", x, y)
				fmt.Println("-> Ready! Press 'Ctrl+Shift+Z' to start the hint bot.")
			}

		case <-hkZ.Keydown():
			if b.HasBotRight {
				b.IsRunning = !b.IsRunning
				if b.IsRunning {
					fmt.Println("[▶] Bot RESUMED")
				} else {
					fmt.Println("[⏸] Bot PAUSED")
					b.CurrentMoves = []Move{}
					b.ActiveMoves[0] = nil // Clear 1st move
					b.ActiveMoves[1] = nil // Clear 2nd move
				}
			}
		}
	}
}

func (b *Bot) Worker() {
	for {
		if !b.IsRunning || !b.HasBotRight {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		currentGrid, _ := b.parseBoard()
		moves := findAllMoves(currentGrid)

		if len(moves) > 0 {
			// Pass 1: Clear out any locked moves that are no longer valid
			for i := 0; i < 2; i++ {
				if b.ActiveMoves[i] != nil {
					valid := false
					for _, m := range moves {
						if m == *b.ActiveMoves[i] {
							valid = true
							break
						}
					}
					if !valid {
						b.ActiveMoves[i] = nil
					}
				}
			}

			// Pass 2: Fill empty slots with unused moves
			for i := 0; i < 2; i++ {
				if b.ActiveMoves[i] == nil {
					// Find the first move that isn't already in another slot
					for moveIdx, m := range moves {
						alreadyUsed := false
						for j := 0; j < 2; j++ {
							if b.ActiveMoves[j] != nil && *b.ActiveMoves[j] == m {
								alreadyUsed = true
								break
							}
						}

						if !alreadyUsed {
							b.ActiveMoves[i] = &moves[moveIdx]
							break
						}
					}
				}
			}
		} else {
			// No moves found on the board at all
			b.ActiveMoves[0] = nil
			b.ActiveMoves[1] = nil
		}

		b.CurrentMoves = moves
		time.Sleep(250 * time.Millisecond)
	}
}
