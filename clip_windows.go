package main

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")
	openClipboard    = user32.NewProc("OpenClipboard")
	closeClipboard   = user32.NewProc("CloseClipboard")
	emptyClipboard   = user32.NewProc("EmptyClipboard")
	setClipboardData = user32.NewProc("SetClipboardData")
	globalAlloc      = kernel32.NewProc("GlobalAlloc")
	globalLock       = kernel32.NewProc("GlobalLock")
	globalUnlock     = kernel32.NewProc("GlobalUnlock")
)

const (
	cfUnicodeText = 13
	gmemMoveable  = 0x0002
)

// copyToClipboard uses the Win32 API directly: works from cmd, PowerShell or anywhere, no helper programs.
func copyToClipboard(text string) error {
	u, err := syscall.UTF16FromString(text)
	if err != nil {
		return err
	}
	if r, _, _ := openClipboard.Call(0); r == 0 {
		return errors.New("cannot open clipboard")
	}
	defer closeClipboard.Call()
	emptyClipboard.Call()
	size := uintptr(len(u) * 2)
	h, _, _ := globalAlloc.Call(gmemMoveable, size)
	if h == 0 {
		return errors.New("alloc failed")
	}
	p, _, _ := globalLock.Call(h)
	if p == 0 {
		return errors.New("lock failed")
	}
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u)), u)
	globalUnlock.Call(h)
	if r, _, _ := setClipboardData.Call(cfUnicodeText, h); r == 0 {
		return errors.New("set clipboard failed")
	}
	return nil
}

// enableVT turns on ANSI escape handling for stdout in cmd/conhost and returns a restore func.
func enableVT() func() {
	h := windows.Handle(windows.Stdout)
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return func() {}
	}
	windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	return func() { windows.SetConsoleMode(h, mode) }
}
