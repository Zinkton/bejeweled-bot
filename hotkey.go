package main

import (
	"fmt"

	"golang.design/x/hotkey"
)

func (g *Game) StartHotkeyListener(bot *BotController) {
	go func() {
		// Ctrl+Shift+H for Hint Mode
		hkHint := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyH)
		if err := hkHint.Register(); err != nil {
			fmt.Println("Failed to register Hint hotkey:", err)
			return
		}
		defer hkHint.Unregister()

		// Ctrl+Shift+B for Bot Mode
		hkBot := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyB)
		if err := hkBot.Register(); err != nil {
			fmt.Println("Failed to register Bot hotkey:", err)
			return
		}
		defer hkBot.Unregister()

		// Ctrl+Shift+F for Bot Speed
		hkBotSpeed := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyS)
		if err := hkBotSpeed.Register(); err != nil {
			fmt.Println("Failed to register Bot hotkey:", err)
			return
		}
		defer hkBotSpeed.Unregister()

		for {
			select {
			case <-hkHint.Keydown():
				g.ToggleMode()
			case <-hkBot.Keydown():
				bot.IsEnabled = !bot.IsEnabled
				g.BotEnabled = bot.IsEnabled
				if bot.IsEnabled {
					fmt.Println("[▶] Autonomous Bot ENABLED")
				} else {
					fmt.Println("[⏸] Autonomous Bot DISABLED")
				}
			case <-hkBotSpeed.Keydown():
				g.FastBotMode = !g.FastBotMode
				if g.FastBotMode {
					fmt.Println("Bot Speed FAST")
				} else {
					fmt.Println("Bot Speed SMART")
				}
			}
		}
	}()
}
