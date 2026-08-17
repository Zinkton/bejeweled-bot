package main

import "time"

type BotController struct {
	IsEnabled bool
	Geom      *WindowBounds
}

func (g *Game) StartBotWorker(bot *BotController) {
	go func() {
		for {
			if !bot.IsEnabled || g.Geom == nil {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			move := g.CachedBestMove
			fastMoves := g.CachedFastMoves
			if move == nil && fastMoves == nil || (move != nil && move.Index1 == move.Index2) || !g.IsActive {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			if move != nil {
				NativeExecuteSwap(g.Geom, move.Index1, move.Index2)
			} else {
				for _, fastMove := range *fastMoves {
					NativeExecuteSwap(g.Geom, fastMove.Index1, fastMove.Index2)
				}
			}

		}
	}()
}
