package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

const (
	targetRed    = 0
	targetOrange = 27
	targetYellow = 53
	targetGreen  = 132
	targetBlue   = 210
	targetPurple = 298
)

func parseBoard() ([8][8]GemColor, []string) {
	var grid [8][8]GemColor
	var logs []string
	sampleSize := 16

	startX := boardAnchorX - (sampleSize / 2)
	startY := boardAnchorY - (sampleSize / 2)
	width := int(float64(7)*stepX) + sampleSize
	height := int(float64(7)*stepY) + sampleSize + 30

	fullBoardImg, _ := robotgo.CaptureImg(startX, startY, width, height)

	for r := range 8 {
		for c := range 8 {
			centerX := boardAnchorX + int(float64(c)*stepX)
			centerY := boardAnchorY + int(float64(r)*stepY) + 15

			detected, logMsg := analyzeGem(fullBoardImg, startX, startY, centerX, centerY, sampleSize, r, c)

			grid[r][c] = detected
			logs = append(logs, logMsg)
		}
	}

	saveDebugData(fullBoardImg, grid, logs)

	return grid, logs
}

func analyzeGem(fullImg image.Image, captureStartX, captureStartY, targetX, targetY, sampleSize, row, col int) (GemColor, string) {
	pixelCount := sampleSize * sampleSize
	imgX := targetX - captureStartX
	imgY := targetY - captureStartY
	minX := imgX - (sampleSize / 2)
	minY := imgY - (sampleSize / 2)
	maxX := minX + sampleSize
	maxY := minY + sampleSize

	counts := make(map[GemColor]int)

	var totalR, totalG, totalB uint64

	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			r, g, b, _ := fullImg.At(x, y).RGBA()
			r8, g8, b8 := int(r>>8), int(g>>8), int(b>>8)

			totalR += uint64(r8)
			totalG += uint64(g8)
			totalB += uint64(b8)

			pixelColor := classifyPixel(r8, g8, b8)
			counts[pixelColor]++
		}
	}

	avgR := int(totalR / uint64(pixelCount))
	avgG := int(totalG / uint64(pixelCount))
	avgB := int(totalB / uint64(pixelCount))
	avgH, avgS, avgV := rgbToHSV(avgR, avgG, avgB)

	var detected GemColor
	var bestGemColor = Empty
	var maxGemCount int
	distinctColors := 0

	noiseThreshold := pixelCount / 8

	for color, count := range counts {
		if color != Empty && count > maxGemCount {
			maxGemCount = count
			bestGemColor = color
		}
		if color != Empty && color != White && count > noiseThreshold {
			distinctColors++
		}
	}

	hasBlack := counts[Empty] > noiseThreshold

	if counts[Empty] > int(float64(pixelCount)*0.75) {
		detected = Empty
	} else if distinctColors >= 3 && hasBlack {
		detected = Hypercube
	} else {
		detected = bestGemColor
	}

	countsDebug := ""
	for c, count := range counts {
		if count > 0 {
			countsDebug += fmt.Sprintf("%s:%d ", c, count)
		}
	}

	logMsg := fmt.Sprintf("Cell [%d][%d] | Result: %-9s | Avg HSV: H:%3d S:%3d V:%3d | Breakdown: %s",
		row, col, detected, avgH, avgS, avgV, countsDebug)

	return detected, logMsg
}

func closestColorByHue(hue int) GemColor {
	bestColor := Empty
	smallestDist := 360

	check := func(targetHue int, color GemColor) {
		dist := int(math.Abs(float64(hue - targetHue)))
		if dist > 180 {
			dist = 360 - dist
		}

		if dist < smallestDist {
			smallestDist = dist
			bestColor = color
		}
	}

	check(targetRed, Red)
	check(targetOrange, Orange)
	check(targetYellow, Yellow)
	check(targetGreen, Green)
	check(targetBlue, Blue)
	check(targetPurple, Purple)

	return bestColor
}

func rgbToHSV(r, g, b int) (hue, sat, val int) {
	R := float64(r) / 255.0
	G := float64(g) / 255.0
	B := float64(b) / 255.0

	max := math.Max(R, math.Max(G, B))
	min := math.Min(R, math.Min(G, B))

	val = int(max * 100)

	if max == 0 {
		sat = 0
	} else {
		sat = int(((max - min) / max) * 100)
	}

	if max == min {
		hue = 0
	} else {
		var h float64
		diff := max - min
		switch max {
		case R:
			h = (G - B) / diff
			if G < B {
				h += 6.0
			}
		case G:
			h = (B-R)/diff + 2.0
		default:
			h = (R-G)/diff + 4.0
		}
		hue = int(h * 60)
	}

	return hue, sat, val
}

func classifyPixel(r, g, b int) GemColor {
	h, s, v := rgbToHSV(r, g, b)

	if v < 45 {
		return Empty
	}

	if s < 40 && v >= 50 {
		return White
	}

	return closestColorByHue(h)
}

func saveDebugData(fullBoardImg image.Image, grid [8][8]GemColor, logs []string) {
	debugDir := "debug_logs"
	os.MkdirAll(debugDir, os.ModePerm)

	timestamp := time.Now().Format("2006-01-02_15-04-05_000")

	imgPath := filepath.Join(debugDir, fmt.Sprintf("board_%s.png", timestamp))
	imgFile, err := os.Create(imgPath)
	if err == nil {
		png.Encode(imgFile, fullBoardImg)
		imgFile.Close()
	}

	txtPath := filepath.Join(debugDir, fmt.Sprintf("board_%s.txt", timestamp))
	txtFile, err := os.Create(txtPath)
	if err == nil {
		defer txtFile.Close()

		txtFile.WriteString("--- Parsed Grid ---\n")
		for _, row := range grid {
			for _, gem := range row {
				txtFile.WriteString(gemToPlainText(gem))
			}
			txtFile.WriteString("\n")
		}
		txtFile.WriteString("-------------------\n\n")

		for _, logMsg := range logs {
			txtFile.WriteString(logMsg + "\n")
		}
	}
}
