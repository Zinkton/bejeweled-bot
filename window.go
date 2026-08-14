package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procFindWindowW    = user32.NewProc("FindWindowW")
	procGetClientRect  = user32.NewProc("GetClientRect")
	procClientToScreen = user32.NewProc("ClientToScreen")
	procGetWindowRect  = user32.NewProc("GetWindowRect")
)

type RECT struct {
	Left, Top, Right, Bottom int32
}

type POINT struct {
	X, Y int32
}

type WindowBounds struct {
	HWND         uintptr
	ScreenX      int // Client area top-left X in screen coordinates
	ScreenY      int // Client area top-left Y in screen coordinates
	ClientWidth  int // Actual width of game canvas
	ClientHeight int // Actual height of game canvas

	// Calculated 8x8 Grid coordinates
	AnchorX int     // Center of Top-Left gem (0,0) on screen
	AnchorY int     // Center of Top-Left gem (0,0) on screen
	StepX   float64 // Horizontal distance between gem centers
	StepY   float64 // Vertical distance between gem centers
}

// FindGameWindow searches for the Bejeweled 3 window by class or window title
func FindGameWindow() (uintptr, error) {
	// Bejeweled 3 typically uses "Bejeweled 3" as its window title or class
	title := "Bejeweled 3"

	ptr, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd != 0 {
		return hwnd, nil
	}

	return 0, fmt.Errorf("bejeweled 3 window not found")
}

// GetGameGeometry queries the window coordinates and calculates the gem grid
func GetGameGeometry() (*WindowBounds, error) {
	hwnd, err := FindGameWindow()
	if err != nil {
		return nil, err
	}

	var clientRect RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&clientRect)))

	clientWidth := int(clientRect.Right - clientRect.Left)
	clientHeight := int(clientRect.Bottom - clientRect.Top)

	// Translate client (0, 0) to absolute screen coordinates
	pt := POINT{X: 0, Y: 0}
	procClientToScreen.Call(hwnd, uintptr(unsafe.Pointer(&pt)))

	bounds := &WindowBounds{
		HWND:         hwnd,
		ScreenX:      int(pt.X),
		ScreenY:      int(pt.Y),
		ClientWidth:  clientWidth,
		ClientHeight: clientHeight,
	}

	bounds.calculateGrid()
	return bounds, nil
}

// calculateGrid determines the exact center of gem (0,0) and spacing
func (wb *WindowBounds) calculateGrid() {
	w := float64(wb.ClientWidth)
	h := float64(wb.ClientHeight)

	// Calibrated ratios for Bejeweled 3 window client area
	tlNormX := 0.3651
	tlNormY := 0.1584
	brNormX := 0.9260
	brNormY := 0.9118

	wb.AnchorX = wb.ScreenX + int(tlNormX*w)
	wb.AnchorY = wb.ScreenY + int(tlNormY*h)

	brX := wb.ScreenX + int(brNormX*w)
	brY := wb.ScreenY + int(brNormY*h)

	wb.StepX = float64(brX-wb.AnchorX) / 7.0
	wb.StepY = float64(brY-wb.AnchorY) / 7.0
}

// GemScreenPos returns the absolute screen X, Y for any grid coordinate (row 0-7, col 0-7)
func (wb *WindowBounds) GemScreenPos(row, col int) (int, int) {
	x := wb.AnchorX + int(float64(col)*wb.StepX)
	y := wb.AnchorY + int(float64(row)*wb.StepY)
	return x, y
}
