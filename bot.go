package main

import "time"

type BotController struct {
	IsEnabled bool
	Geom      *WindowBounds
}

func (g *Game) StartBotWorker(bot *BotController) {
	go func() {
		for {
			// 1. Skip if bot is disabled or geometry isn't calibrated
			if !bot.IsEnabled || bot.Geom == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 2. Skip if the game is paused (timer not ticking)
			if !g.IsActive {
				time.Sleep(20 * time.Millisecond)
				continue
			}

			// 3. Grab current cached best move
			move := g.CachedBestMove
			if move == nil || move.Index1 == move.Index2 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// 4. Execute the move
			NativeExecuteSwap(bot.Geom, move.Index1, move.Index2)
		}
	}()
}
