package kioskweb

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output user32.go tools.go

//sys EnumWindows(lpEnumFunc uintptr, lParam uintptr) (err error) = user32.EnumWindows
//sys GetWindowTextW(hwnd syscall.Handle, text *uint16, nMaxCount int32) (err error) = user32.GetWindowTextW
//sys SetForegroundWindow(hwnd syscall.Handle) (err error) = user32.SetForegroundWindow
