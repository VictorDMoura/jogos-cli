//go:build linux

package engine

import (
	"os"
	"syscall"
	"unsafe"
)

type platformState struct {
	fd      uintptr
	termios syscall.Termios
}

func enablePlatformRawMode() (*platformState, error) {
	fd := os.Stdin.Fd()
	var orig syscall.Termios
	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, uintptr(unsafe.Pointer(&orig)), 0, 0, 0); errno != 0 {
		return nil, errno
	}

	raw := orig
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0

	if _, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); errno != 0 {
		return nil, errno
	}

	return &platformState{fd: fd, termios: orig}, nil
}

func restorePlatformRawMode(st *platformState) {
	if st == nil {
		return
	}
	_, _, _ = syscall.Syscall6(syscall.SYS_IOCTL, st.fd, syscall.TCSETS, uintptr(unsafe.Pointer(&st.termios)), 0, 0, 0)
}
