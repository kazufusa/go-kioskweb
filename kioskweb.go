// +build windows

package kioskweb

import (
	"context"
	"net/http"
	"os/exec"
	"time"
)

type Browser string

const (
	IE   Browser = "IE"
	Edge Browser = "Edge"
	// Chrome
	// Firefox
)

var (
	args = map[Browser][]string{
		IE:   {"/c", "start", "iexplore.exe", "-k"},
		Edge: {"/c", "start", "msedge.exe", "--kiosk", "--edge-kiosk-type=fullscreen"},
	}
)

type Config struct {
	Browser Browser
	WaitCtx context.Context
}

func OpenKioskWeb(url string, config *Config) error {
	if config.WaitCtx != nil {
		err := wait(config.WaitCtx, url)
		if err != nil {
			return err
		}
	}

	_args := append([]string{}, args[config.Browser]...)
	_args = append(_args, url)
	return exec.Command("cmd", _args...).Run()
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
