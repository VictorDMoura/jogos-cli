//go:build windows

package engine

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	enableProcessedInput        = 0x0001
	enableLineInput             = 0x0002
	enableEchoInput             = 0x0004
	enableWindowInput           = 0x0008
	enableVirtualTerminalInput  = 0x0200
	enableProcessedOutput       = 0x0001
	enableVirtualTerminalOutput = 0x0004
	utf8CodePage                = 65001
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode       = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode       = kernel32.NewProc("SetConsoleMode")
	procGetConsoleOutputCP   = kernel32.NewProc("GetConsoleOutputCP")
	procSetConsoleOutputCP   = kernel32.NewProc("SetConsoleOutputCP")
)

type platformState struct {
	inHandle  syscall.Handle
	outHandle syscall.Handle
	inMode    uint32
	outMode   uint32
	outCP     uintptr
}

func getConsoleMode(handle syscall.Handle) (uint32, error) {
	var mode uint32
	r1, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return 0, err
	}
	return mode, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	r1, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if r1 == 0 {
		return err
	}
	return nil
}

func enablePlatformRawMode() (*platformState, error) {
	inHandle := syscall.Handle(os.Stdin.Fd())
	outHandle := syscall.Handle(os.Stdout.Fd())

	inMode, err := getConsoleMode(inHandle)
	if err != nil {
		return nil, err
	}
	outMode, err := getConsoleMode(outHandle)
	if err != nil {
		return nil, err
	}

	origCP, _, _ := procGetConsoleOutputCP.Call()
	procSetConsoleOutputCP.Call(uintptr(utf8CodePage))

	// Desativa line input, echo e processed input; ativa Virtual Terminal Input para setas ANSI
	newInMode := inMode &^ (enableEchoInput | enableLineInput | enableProcessedInput)
	newInMode |= enableVirtualTerminalInput
	if err := setConsoleMode(inHandle, newInMode); err != nil {
		return nil, err
	}

	// Ativa sequências de escape ANSI na saída
	newOutMode := outMode | enableProcessedOutput | enableVirtualTerminalOutput
	_ = setConsoleMode(outHandle, newOutMode)

	return &platformState{
		inHandle:  inHandle,
		outHandle: outHandle,
		inMode:    inMode,
		outMode:   outMode,
		outCP:     origCP,
	}, nil
}

func restorePlatformRawMode(st *platformState) {
	if st == nil {
		return
	}
	_ = setConsoleMode(st.inHandle, st.inMode)
	_ = setConsoleMode(st.outHandle, st.outMode)
	if st.outCP != 0 {
		procSetConsoleOutputCP.Call(st.outCP)
	}
}
