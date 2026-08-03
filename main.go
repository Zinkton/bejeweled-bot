package main

import (
	"fmt"

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

func executeBotCycle() {
	if !hasTopLeft || !hasBotRight {
		fmt.Println("\n[Error] Please calibrate BOTH corners of the board first!")
		return
	}

	fmt.Println("\n--- Bot Cycle Started ---")

	fmt.Println("1. Capturing and parsing screen data...")
	currentGrid := parseBoard()

	fmt.Println("Parsed Board:")
	printGrid(currentGrid)

	fmt.Println("2. Calculating move...")
	move, found := findHint(currentGrid)

	if found {
		fmt.Printf("   -> Move found! Swap Row %d, Col %d to the %s.\n", move.Row, move.Col, move.Dir)
		executeMove(move)
	} else {
		fmt.Println("   -> No valid moves found on this board.")
	}
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

func main() {
	fmt.Println("Bejeweled Bot Initialized!")
	fmt.Println("1. Hover Top-Left gem and press 'Ctrl + Shift + X'")
	fmt.Println("2. Hover Bottom-Right gem and press 'Ctrl + Shift + C'")
	fmt.Println("3. Press 'Ctrl + Shift + Z' to trigger a bot cycle.")
	fmt.Println("Press 'Ctrl + C' in this terminal to quit.")

	go listenForTopLeft()
	go listenForBottomRight()

	botHk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyZ)
	botHk.Register()
	defer botHk.Unregister()

	for {
		<-botHk.Keydown()
		executeBotCycle()
	}
}
