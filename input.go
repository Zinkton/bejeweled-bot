package main

import (
	"time"
	"unsafe"
)

var (
	procSetCursorPos = user32.NewProc("SetCursorPos")
	procGetCursorPos = user32.NewProc("GetCursorPos")
	procMouseEvent   = user32.NewProc("mouse_event")
	procKeybdEvent   = user32.NewProc("keybd_event")
)

const (
	MOUSEEVENTF_LEFTDOWN = 0x0002
	MOUSEEVENTF_LEFTUP   = 0x0004

	VK_LEFT  = 0x25
	VK_UP    = 0x26
	VK_RIGHT = 0x27
	VK_DOWN  = 0x28

	KEYEVENTF_KEYUP = 0x0002
)

// MoveMouse moves the cursor directly in Windows Virtual Screen coordinates (supports negative X/Y)
func MoveMouse(x, y int) {
	procSetCursorPos.Call(uintptr(int32(x)), uintptr(int32(y)))
}

// GetMousePos returns the current mouse location in physical pixels
func GetMousePos() (int, int) {
	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

// IndexToRowCol converts a 0-63 index into grid row and column
func IndexToRowCol(index int) (row, col int) {
	return index / 8, index % 8
}

// GetKeyDirection figures out which arrow key to press
func GetKeyDirection(idx1, idx2 int) (string, bool) {
	diff := idx2 - idx1
	switch diff {
	case 1:
		return "right", true
	case -1:
		return "left", true
	case 8:
		return "down", true
	case -8:
		return "up", true
	default:
		return "", false // Not adjacent or identical
	}
}

// Click sends a native left-click
func Click() {
	procMouseEvent.Call(MOUSEEVENTF_LEFTDOWN, 0, 0, 0, 0)
	time.Sleep(20 * time.Millisecond)
	procMouseEvent.Call(MOUSEEVENTF_LEFTUP, 0, 0, 0, 0)
}

// TapKey sends a native directional arrow keypress
func TapKey(dir string) {
	var vk uintptr
	switch dir {
	case "left":
		vk = VK_LEFT
	case "up":
		vk = VK_UP
	case "right":
		vk = VK_RIGHT
	case "down":
		vk = VK_DOWN
	default:
		return
	}

	procKeybdEvent.Call(vk, 0, 0, 0)
	time.Sleep(20 * time.Millisecond)
	procKeybdEvent.Call(vk, 0, KEYEVENTF_KEYUP, 0)
}

// NativeExecuteSwap clicks the source gem and presses the direction key
func NativeExecuteSwap(geom *WindowBounds, idx1, idx2 int) {
	if idx1 == idx2 || geom == nil {
		return
	}

	key, ok := GetKeyDirection(idx1, idx2)
	if !ok {
		return
	}

	r, c := IndexToRowCol(idx1)
	screenX, screenY := geom.GemScreenPos(r, c)

	MoveMouse(screenX, screenY)
	//time.Sleep(15 * time.Millisecond)
	Click()
	//time.Sleep(15 * time.Millisecond)
	TapKey(key)
}
