//go:build darwin

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

func enableRawMode() {
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGETA),
		uintptr(unsafe.Pointer(&origTermios)),
		0, 0, 0,
	)
	if errno != 0 {
		return
	}
	raw := origTermios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCSETA),
		uintptr(unsafe.Pointer(&raw)),
		0, 0, 0,
	)
}

func disableRawMode() {
	syscall.Syscall6(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCSETA),
		uintptr(unsafe.Pointer(&origTermios)),
		0, 0, 0,
	)
}
