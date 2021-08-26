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

type Browser string

const (
	IE      Browser = "IE"
	Edge    Browser = "Edge"
	Chrome  Browser = "Chrome"
	Firefox Browser = "Firefox"
)

var (
	args = map[Browser][]string{
		IE:     {"/c", "start", "iexplore.exe", "-k"},
		Edge:   {"/c", "start", "msedge.exe", "--kiosk", "--edge-kiosk-type=fullscreen"},
		Chrome: {"/c", "start", "chrome.exe", "--kiosk", "--disable-pinch"},
	}

	titles = map[Browser]*regexp.Regexp{
		Chrome:  regexp.MustCompile(`- Google Chrome$`),
		IE:      regexp.MustCompile(`- Internet Explorer$`),
		Edge:    regexp.MustCompile(`- Microsoft​ Edge$`),
		Firefox: regexp.MustCompile(`— Mozilla Firefox$`),
	}
)

type Config struct {
	Browser Browser
	WaitCtx context.Context
}

func OpenKioskWeb(url string, config *Config) error {
	pHandles, err := FindWindows(titles[config.Browser])
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
	if len(pHandles) != 0 {
		for i := 0; i < 10; i++ {
			handles, err = FindWindows(titles[config.Browser])
			if err != nil {
				return err
			}
			handle, err = NewlyOpenedWindow(pHandles, handles)
			if err != nil {
				time.Sleep(100 * time.Millisecond)
			} else {
				break
			}
		}
		if err != nil {
			return err
		}
	} else {
		for i := 0; i < 10; i++ {
			handles, err = FindWindows(titles[config.Browser])
			if err != nil {
				return err
			}
			if len(handles) > 0 {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if len(handles) == 0 {
			return errors.New("not found")
		}
		handle = handles[0]
	}

	// escape from start menu in tablet mode
	err = pressAltTab()
	if err != nil {
		return err
	}

	return SetForegroundWindow(handle)
}

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

// FindWindows finds all currently opened browsers
func FindWindows(reTitle *regexp.Regexp) (ret []syscall.Handle, err error) {
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

func NewlyOpenedWindow(previous, current []syscall.Handle) (ret syscall.Handle, err error) {
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
