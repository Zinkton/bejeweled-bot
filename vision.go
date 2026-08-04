package main

import (
	"fmt"
	"image"
	"math"

	"github.com/go-vgo/robotgo"
)

const (
	targetRed    = 0
	targetOrange = 30
	targetYellow = 58
	targetGreen  = 130
	targetBlue   = 210
	targetPurple = 300
)

func parseBoard() ([8][8]GemColor, []string) {
	var grid [8][8]GemColor
	var logs []string
	sampleSize := 10
	yOffset := 15

	for r := range 8 {
		for c := range 8 {
			centerX := boardAnchorX + int(float64(c)*stepX)
			centerY := boardAnchorY + int(float64(r)*stepY)

			captureY := centerY + yOffset

			img, _ := robotgo.CaptureImg(centerX-(sampleSize/2), captureY-(sampleSize/2), sampleSize, sampleSize)

			detected, logMsg := analyzeGem(img, r, c)

			grid[r][c] = detected
			logs = append(logs, logMsg)
		}
	}

	return grid, logs
}

func analyzeGem(img image.Image, row, col int) (GemColor, string) {
	bounds := img.Bounds()
	var totalR, totalG, totalB uint64
	pixelCount := uint64(bounds.Dx() * bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			totalR += uint64(r >> 8)
			totalG += uint64(g >> 8)
			totalB += uint64(b >> 8)
		}
	}

	avgR := int(totalR / pixelCount)
	avgG := int(totalG / pixelCount)
	avgB := int(totalB / pixelCount)

	h, s, v := rgbToHSV(avgR, avgG, avgB)

	var detected GemColor

	if v < 15 {
		detected = Empty
	} else if v >= 15 && v < 45 {
		detected = Hypercube
	} else if s < 15 && v >= 50 {
		detected = White
	} else {
		detected = closestColorByHue(h)
	}

	// Format the log message but DO NOT print it yet
	logMsg := fmt.Sprintf("Cell [%d][%d] | RGB: (%3d, %3d, %3d) | HSV: H:%3d S:%3d V:%3d | Result: %s",
		row, col, avgR, avgG, avgB, h, s, v, detected)

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
