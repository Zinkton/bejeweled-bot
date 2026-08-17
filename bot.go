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
			if !bot.IsEnabled || g.Geom == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			// 3. Grab current cached best move
			moves := g.CachedBestMove
			if len(moves) == 0 || !g.IsActive {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// 4. Execute the move
			for _, move := range moves {
				NativeExecuteSwap(g.Geom, move.Index1, move.Index2)
			}

		}
	}()
}
