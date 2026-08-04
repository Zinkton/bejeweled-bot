package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"
)

var (
	boardAnchorX int
	boardAnchorY int

	stepX float64
	stepY float64

	hasTopLeft  bool
	hasBotRight bool

	botRunning bool
)

func calibrateTopLeft() {
	x, y := robotgo.Location()
	boardAnchorX = x
	boardAnchorY = y
	hasTopLeft = true
	fmt.Printf("\n[Calibrated] Top-Left Gem anchored at X: %d, Y: %d\n", x, y)
}

func calibrateBottomRight() {
	if !hasTopLeft {
		fmt.Println("\n[Warning] Please calibrate the Top-Left gem first!")
		return
	}

	x, y := robotgo.Location()

	stepX = float64(x-boardAnchorX) / 7.0
	stepY = float64(y-boardAnchorY) / 7.0
	hasBotRight = true

	fmt.Printf("\n[Calibrated] Bottom-Right Gem anchored at X: %d, Y: %d\n", x, y)
	fmt.Printf("[Success] Grid spacing calculated -> X: %.2fpx, Y: %.2fpx\n", stepX, stepY)
}

func executeMove(m Move) {
	screenX := boardAnchorX + int(float64(m.Col)*stepX)
	screenY := boardAnchorY + int(float64(m.Row)*stepY)

	robotgo.Move(screenX, screenY)
	robotgo.Click("left")

	switch m.Dir {
	case "Right":
		robotgo.KeyTap("right")
	case "Down":
		robotgo.KeyTap("down")
	case "Left":
		robotgo.KeyTap("left")
	case "Up":
		robotgo.KeyTap("up")
	}
}

func listenForTopLeft() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyX)
	hk.Register()
	defer hk.Unregister()

	for {
		<-hk.Keydown()
		calibrateTopLeft()
	}
}

func listenForBottomRight() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyC)
	hk.Register()
	defer hk.Unregister()

	for {
		<-hk.Keydown()
		calibrateBottomRight()
	}
}

func toggleBot() {
	if !hasTopLeft || !hasBotRight {
		fmt.Println("\n[Error] Please calibrate BOTH corners of the board first!")
		return
	}

	// currentGrid, logs := parseBoard()
	// printGrid(currentGrid)
	// fmt.Println("\n--- Vision Debug Log ---")
	// for _, l := range logs {
	// 	fmt.Println(l)
	// }
	// fmt.Println("------------------------")
	// return

	botRunning = !botRunning

	if botRunning {
		fmt.Println("\n[▶] Bot STARTED! Press 'Ctrl + Shift + Z' to pause.")
	} else {
		fmt.Println("\n[⏸] Bot PAUSED! Press 'Ctrl + Shift + Z' to resume.")
	}
}

func botWorker() {
	for {
		if botRunning {
			// fmt.Println("\n--- New Turn ---")
			currentGrid, _ := parseBoard()

			moves := findAllMoves(currentGrid)

			if len(moves) > 0 {
				// fmt.Printf("   -> Found %d potential moves! Executing queue...\n", len(moves))

				usedCells := make(map[string]bool)
				executedCount := 0

				for _, m := range moves {
					targetID := fmt.Sprintf("%d,%d", m.Row, m.Col)

					if usedCells[targetID] {
						continue
					}

					// fmt.Printf("      * Executing: Swap Row %d, Col %d to the %s.\n", m.Row, m.Col, m.Dir)
					executeMove(m)
					executedCount++

					usedCells[targetID] = true

					time.Sleep(75 * time.Millisecond)
				}

				// fmt.Printf("   -> Queue finished (%d moves executed). Waiting for board to settle...\n", executedCount)

				time.Sleep(350 * time.Millisecond)
			} else {
				// fmt.Println("   -> [ERROR] No valid moves found! Dumping vision data...")

				// fmt.Println("\n--- Vision Debug Log ---")
				// for _, l := range logs {
				// 	fmt.Println(l)
				// }
				// fmt.Println("------------------------")

				// printGrid(currentGrid)

				time.Sleep(75 * time.Millisecond)
			}
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func main() {
	fmt.Println("Bejeweled Bot Initialized!")
	fmt.Println("1. Hover Top-Left gem and press 'Ctrl + Shift + X'")
	fmt.Println("2. Hover Bottom-Right gem and press 'Ctrl + Shift + C'")
	fmt.Println("3. Press 'Ctrl + Shift + Z' to toggle the bot ON/OFF.")
	fmt.Println("Press 'Ctrl + C' in this terminal to quit completely.")

	go listenForTopLeft()
	go listenForBottomRight()

	go botWorker()

	botHk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyZ)
	botHk.Register()
	defer botHk.Unregister()

	for {
		<-botHk.Keydown()
		toggleBot()
	}
}
