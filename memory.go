package main

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

const (
	PROCESS_VM_READ           = 0x0010
	PROCESS_QUERY_INFORMATION = 0x0400

	TH32CS_SNAPPROCESS  = 0x00000002
	TH32CS_SNAPMODULE   = 0x00000008
	TH32CS_SNAPMODULE32 = 0x00000010
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procReadProcessMemory        = kernel32.NewProc("ReadProcessMemory")
	procOpenProcess              = kernel32.NewProc("OpenProcess")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procModule32FirstW           = kernel32.NewProc("Module32FirstW")
)

// Windows API Structs
type PROCESSENTRY32 struct {
	Size              uint32
	CntUsage          uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	CntThreads        uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [260]uint16
}

type MODULEENTRY32 struct {
	Size         uint32
	ModuleID     uint32
	ProcessID    uint32
	GlblcntUsage uint32
	ProccntUsage uint32
	ModBaseAddr  uintptr
	ModBaseSize  uint32
	ModuleHandle uintptr
	Module       [256]uint16
	ExePath      [260]uint16
}

var processHandle syscall.Handle
var baseAddress uint32 // Bejeweled 3 is 32-bit

// AttachToProcess finds the PID by name, opens it, and gets the base memory address.
func AttachToProcess(exeName string) bool {
	// 1. Snapshot all processes
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(syscall.InvalidHandle) {
		return false
	}
	defer procCloseHandle.Call(snapshot)

	var pe32 PROCESSENTRY32
	pe32.Size = uint32(unsafe.Sizeof(pe32))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&pe32)))
	if ret == 0 {
		return false
	}

	var pid uint32 = 0
	// Loop through processes to find ours
	for {
		name := syscall.UTF16ToString(pe32.ExeFile[:])
		if strings.EqualFold(name, exeName) {
			pid = pe32.ProcessID
			break
		}
		ret, _, _ := procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&pe32)))
		if ret == 0 {
			break
		}
	}

	if pid == 0 {
		return false
	}

	// 2. Open the process
	handle, _, _ := procOpenProcess.Call(PROCESS_VM_READ|PROCESS_QUERY_INFORMATION, 0, uintptr(pid))
	if handle == 0 {
		return false
	}

	// 3. Snapshot modules to get the Base Address
	modSnapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPMODULE|TH32CS_SNAPMODULE32, uintptr(pid))
	if modSnapshot == uintptr(syscall.InvalidHandle) {
		procCloseHandle.Call(handle)
		return false
	}
	defer procCloseHandle.Call(modSnapshot)

	var me32 MODULEENTRY32
	me32.Size = uint32(unsafe.Sizeof(me32))

	ret, _, _ = procModule32FirstW.Call(modSnapshot, uintptr(unsafe.Pointer(&me32)))
	if ret != 0 {
		// The first module in the snapshot is always the main executable itself
		baseAddress = uint32(me32.ModBaseAddr)
		processHandle = syscall.Handle(handle)
		return true
	}

	procCloseHandle.Call(handle)
	return false
}

// ReadUint8 reads 1 byte from memory
func ReadUint8(address uint32) (uint8, error) {
	var value uint8
	var bytesRead uintptr

	ret, _, err := procReadProcessMemory.Call(
		uintptr(processHandle),
		uintptr(address),
		uintptr(unsafe.Pointer(&value)),
		1, // Size: 1 byte
		uintptr(unsafe.Pointer(&bytesRead)),
	)

	if ret == 0 {
		return 0, fmt.Errorf("read failed: %v", err)
	}
	return value, nil
}

// ReadUint32 reads 4 bytes from memory
func ReadUint32(address uint32) (uint32, error) {
	var value uint32
	var bytesRead uintptr

	ret, _, err := procReadProcessMemory.Call(
		uintptr(processHandle),
		uintptr(address),
		uintptr(unsafe.Pointer(&value)),
		4,
		uintptr(unsafe.Pointer(&bytesRead)),
	)

	if ret == 0 {
		return 0, fmt.Errorf("read failed: %v", err)
	}
	return value, nil
}
