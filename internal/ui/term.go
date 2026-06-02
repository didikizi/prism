package ui

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// IsTTY reports whether f is connected to an interactive terminal. When it is
// not (a pipe, CI log, or redirect) the live spinner is suppressed.
func IsTTY(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

type winsize struct{ row, col, x, y uint16 }

// Width returns the terminal column count, falling back to $COLUMNS then 100.
func Width(f *os.File) int {
	var ws winsize
	if _, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		f.Fd(),
		syscall.TIOCGWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	); e == 0 && ws.col > 0 {
		return int(ws.col)
	}
	if s := os.Getenv("COLUMNS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 100
}
