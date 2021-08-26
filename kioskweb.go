// +build windows

package kioskweb

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output user32.go kioskweb.go

//sys EnumWindows(lpEnumFunc uintptr, lParam uintptr) (err error) = user32.EnumWindows
//sys GetWindowTextW(hwnd syscall.Handle, text *uint16, nMaxCount int32) (err error) = user32.GetWindowTextW
//sys SetForegroundWindow(hwnd syscall.Handle) (err error) = user32.SetForegroundWindow

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"regexp"
	"syscall"
	"time"

	"github.com/micmonay/keybd_event"
)

type browser string

const (
	IE      browser = "IE"
	Edge    browser = "Edge"
	Chrome  browser = "Chrome"
	Firefox browser = "Firefox"
)

var (
	args = map[browser][]string{
		IE:      {"/c", "start", "iexplore.exe", "-k"},
		Edge:    {"/c", "start", "msedge.exe", "--kiosk", "--edge-kiosk-type=fullscreen", "--new-window"},
		Chrome:  {"/c", "start", "chrome.exe", "--new-window", "--kiosk", "--disable-pinch", "--user-data-dir=%TMP%/kioskweb"},
		Firefox: {"/c", "start", "firefox.exe", "--kiosk", "--new-window"},
	}

	titleRegExps = map[browser]*regexp.Regexp{
		IE:      regexp.MustCompile(`- Internet Explorer$`),
		Edge:    regexp.MustCompile(`- Microsoft​ Edge$`),
		Chrome:  regexp.MustCompile(`- Google Chrome$`),
		Firefox: regexp.MustCompile(`— Mozilla Firefox$`),
	}
)

type Config struct {
	Browser browser
	WaitCtx context.Context
}

// OpenKiosk open url with the user selected browser which is in kiosk mode
func OpenKioskWeb(url string, config Config) error {
	pHandles, err := findWindows(titleRegExps[config.Browser])
	if err != nil {
		return err
	}

	if config.WaitCtx != nil {
		err := wait(config.WaitCtx, url)
		if err != nil {
			return err
		}
	}

	_args := append([]string{}, args[config.Browser]...)
	_args = append(_args, url)
	err = exec.Command("cmd", _args...).Run()
	if err != nil {
		return err
	}

	var handle syscall.Handle
	var handles []syscall.Handle

	for i := 0; i < 20; i++ {
		handles, err = findWindows(titleRegExps[config.Browser])
		if err != nil {
			return err
		}

		if len(pHandles) > 0 {
			handle, err = newlyOpenedWindow(pHandles, handles)
			if err == nil {
				break
			}
		} else {
			if len(handles) > 0 {
				handle = handles[0]
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// escape from start menu if Windows is in tablet mode
	_ = pressAltTab()
	time.Sleep(200 * time.Millisecond)

	if (int)(handle) == 0 {
		return errors.New("not found")
	}
	return SetForegroundWindow(handle)
}

// wait waits for a url to become available.
func wait(ctx context.Context, url string) error {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			_, err := http.Get(url)
			if err == nil {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// findWindows returns the syscall.Handle of all current opened application windows.
func findWindows(reTitle *regexp.Regexp) (ret []syscall.Handle, err error) {
	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		bytes := make([]uint16, 256)
		err := GetWindowTextW(h, &bytes[0], int32(len(bytes)))
		if err == nil {
			title := syscall.UTF16ToString(bytes)
			if reTitle.MatchString(title) {
				ret = append(ret, h)
			}
		}
		return 1 // continue enumeration
	})

	err = EnumWindows(cb, 0)
	return ret, err
}

// newlyOpenedWindow returns a syscall.Handle newly opened.
func newlyOpenedWindow(previous, current []syscall.Handle) (ret syscall.Handle, err error) {
	for _, c := range current {
		for _, p := range previous {
			if c == p {
				goto CONTINUE
			}
		}
		return c, nil
	CONTINUE:
	}
	return ret, errors.New("not found")
}

// pressAltTab emits a keyboard event of ALT+TAB.
func pressAltTab() error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		return err
	}
	kb.SetKeys(keybd_event.VK_TAB)
	kb.HasALT(true)
	err = kb.Launching()
	return err
}
