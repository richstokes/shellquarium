//go:build linux

package main

import (
	"syscall"
	"unsafe"
)

var origTermios syscall.Termios

func getTerminalSize() (int, int) {
	type winsize struct{ Row, Col, Xpx, Ypx uint16 }
	ws := &winsize{}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if errno == 0 && ws.Col > 0 && ws.Row > 0 {
		return int(ws.Col), int(ws.Row)
	}
	return 80, 24
}

const (
	tcgets = 0x5401
	tcsets = 0x5402
)

func enableRawMode() {
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tcgets),
		uintptr(unsafe.Pointer(&origTermios)),
	)
	if errno != 0 {
		return
	}
	raw := origTermios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tcsets),
		uintptr(unsafe.Pointer(&raw)),
	)
}

func disableRawMode() {
	syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(tcsets),
		uintptr(unsafe.Pointer(&origTermios)),
	)
}
