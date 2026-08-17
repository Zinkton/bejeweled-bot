package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	WDA_NONE               = 0x00000000
	WDA_EXCLUDEFROMCAPTURE = 0x00000011
	HWND_TOPMOST           = ^uintptr(0) // -1
	SWP_NOACTIVATE         = 0x0010
	SWP_SHOWWINDOW         = 0x0040
)

var (
	procSetAffinity  = user32.NewProc("SetWindowDisplayAffinity")
	procSetWindowPos = user32.NewProc("SetWindowPos")
)

const OverlayWindowTitle = "Bejeweled Bot - Overlay"

// SetOverlayCloaked prevents the overlay from showing in screenshots/screen recordings
func SetOverlayCloaked(title string, enable bool) {
	ptr, _ := syscall.UTF16PtrFromString(title)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd == 0 {
		return
	}

	affinity := WDA_NONE
	if enable {
		affinity = WDA_EXCLUDEFROMCAPTURE
	}
	procSetAffinity.Call(hwnd, uintptr(affinity))
}

func snapEbitenWindow(x, y, width, height int) {
	ptr, _ := syscall.UTF16PtrFromString(OverlayWindowTitle)
	hwnd, _, _ := procFindWindowW.Call(0, uintptr(unsafe.Pointer(ptr)))
	if hwnd == 0 {
		fmt.Println("[!] Could not find Ebiten overlay HWND to snap")
		return
	}

	procSetWindowPos.Call(
		hwnd,
		HWND_TOPMOST,
		uintptr(int32(x)),
		uintptr(int32(y)),
		uintptr(int32(width)),
		uintptr(int32(height)),
		SWP_NOACTIVATE|SWP_SHOWWINDOW,
	)
}

func (g *Game) ToggleMode() {
	if g.Mode == ModeDebug {
		geom, err := GetGameGeometry()
		if err != nil {
			fmt.Println("[Debug-Error] Could not locate Bejeweled 3:", err)
			return
		}
		g.Geom = geom

		fmt.Printf("\n=== Snapping Overlay (Native Win32) ===\n")
		fmt.Printf("Physical Target Pos:  (%d, %d)\n", geom.ScreenX, geom.ScreenY)
		fmt.Printf("Physical Target Size: %dx%d\n", geom.ClientWidth, geom.ClientHeight)

		// 1. Remove borders and enable click-through
		ebiten.SetWindowDecorated(false)
		ebiten.SetWindowMousePassthrough(true)
		ebiten.SetWindowFloating(true)

		// 2. Snap using native Win32 (bypasses GLFW/Ebiten multi-monitor DPI translation)
		snapEbitenWindow(geom.ScreenX, geom.ScreenY, geom.ClientWidth, geom.ClientHeight)

		// 3. Cloak window from capture
		SetOverlayCloaked(OverlayWindowTitle, true)

		g.Mode = ModeHintOverlay
		fmt.Println("[✔] Mode switched to Hint Overlay")
	} else {
		// Restore normal debug window
		ebiten.SetWindowDecorated(true)
		ebiten.SetWindowMousePassthrough(false)
		ebiten.SetWindowFloating(false)
		ebiten.SetWindowSize(640, 480)

		SetOverlayCloaked(OverlayWindowTitle, false)

		g.Mode = ModeDebug
		fmt.Println("[✔] Mode switched to Debug View")
	}
}

func (g *Game) drawHintOverlay(screen *ebiten.Image) {
	return
}
