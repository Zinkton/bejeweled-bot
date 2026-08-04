package main

import (
	"fmt"
	"time"

	"github.com/go-vgo/robotgo"
	"golang.design/x/hotkey"
)

type Bot struct {
	AnchorX int
	AnchorY int
	StepX   float64
	StepY   float64

	HasTopLeft  bool
	HasBotRight bool
	IsRunning   bool

	MoveHistory []string
}

func main() {
	fmt.Println("Bejeweled Bot Initialized!")
	fmt.Println("1. Hover Top-Left gem and press 'Ctrl + Shift + X'")
	fmt.Println("2. Hover Bottom-Right gem and press 'Ctrl + Shift + C'")
	fmt.Println("3. Press 'Ctrl + Shift + Z' to toggle the bot ON/OFF.")
	fmt.Println("Press 'Ctrl + C' in this terminal to quit completely.")

	bot := &Bot{}

	go bot.ListenForTopLeft()
	go bot.ListenForBottomRight()
	go bot.Worker()

	botHk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyZ)
	botHk.Register()
	defer botHk.Unregister()

	for {
		<-botHk.Keydown()
		bot.Toggle()
	}
}

func (b *Bot) ListenForTopLeft() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyX)
	hk.Register()
	defer hk.Unregister()

	for {
		<-hk.Keydown()
		x, y := robotgo.Location()
		b.AnchorX = x
		b.AnchorY = y
		b.HasTopLeft = true
		fmt.Printf("\n[Calibrated] Top-Left Gem anchored at X: %d, Y: %d\n", x, y)
	}
}

func (b *Bot) ListenForBottomRight() {
	hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyC)
	hk.Register()
	defer hk.Unregister()

	for {
		<-hk.Keydown()
		if !b.HasTopLeft {
			fmt.Println("\n[Warning] Please calibrate the Top-Left gem first!")
			continue
		}

		x, y := robotgo.Location()
		b.StepX = float64(x-b.AnchorX) / 7.0
		b.StepY = float64(y-b.AnchorY) / 7.0
		b.HasBotRight = true

		fmt.Printf("\n[Calibrated] Bottom-Right Gem anchored at X: %d, Y: %d\n", x, y)
		fmt.Printf("[Success] Grid spacing calculated -> X: %.2fpx, Y: %.2fpx\n", b.StepX, b.StepY)
	}
}

func (b *Bot) Toggle() {
	if !b.HasTopLeft || !b.HasBotRight {
		fmt.Println("\n[Error] Please calibrate BOTH corners of the board first!")
		return
	}

	b.IsRunning = !b.IsRunning

	if b.IsRunning {
		fmt.Println("\n[▶] Bot STARTED! Press 'Ctrl + Shift + Z' to pause.")
		b.MoveHistory = nil
	} else {
		fmt.Println("\n[⏸] Bot PAUSED! Press 'Ctrl + Shift + Z' to resume.")
	}
}

func (b *Bot) Worker() {
	for {
		if !b.IsRunning {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		currentGrid, _ := b.parseBoard()
		moves := findAllMoves(currentGrid)

		if len(moves) == 0 {
			b.MoveHistory = nil
			time.Sleep(10 * time.Millisecond)
			continue
		}

		usedCells := make(map[string]bool)
		cycleDetected := false

		for _, m := range moves {
			targetID := fmt.Sprintf("%d,%d", m.Row, m.Col)
			if usedCells[targetID] {
				continue
			}

			b.ExecuteMove(m)
			usedCells[targetID] = true

			moveStr := fmt.Sprintf("%d,%d,%s", m.Row, m.Col, m.Dir)

			if b.trackAndCheckCycle(moveStr) {
				cycleDetected = true
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		if cycleDetected {
			fmt.Println("   -> [!] Phantom cycle detected! Waiting 500ms for board to snap back...")
			time.Sleep(500 * time.Millisecond)
			b.MoveHistory = nil
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (b *Bot) ExecuteMove(m Move) {
	screenX := b.AnchorX + int(float64(m.Col)*b.StepX)
	screenY := b.AnchorY + int(float64(m.Row)*b.StepY)

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

func (b *Bot) trackAndCheckCycle(move string) bool {
	b.MoveHistory = append(b.MoveHistory, move)

	if len(b.MoveHistory) > 20 {
		b.MoveHistory = b.MoveHistory[len(b.MoveHistory)-20:]
	}

	n := len(b.MoveHistory)
	maxCycleLen := 5

	for k := 1; k <= maxCycleLen; k++ {
		if n >= 2*k {
			chunk1 := b.MoveHistory[n-k : n]
			chunk2 := b.MoveHistory[n-2*k : n-k]

			isMatch := true
			for i := range chunk1 {
				if chunk1[i] != chunk2[i] {
					isMatch = false
					break
				}
			}

			if isMatch {
				return true
			}
		}
	}

	return false
}
