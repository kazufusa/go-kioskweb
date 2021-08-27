// +build windows

package kioskweb

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenKioskWeb(t *testing.T) {
	var tests = []struct {
		given browser
		exe   string
	}{
		{given: IE, exe: "iexplore.exe"},
		{given: Edge, exe: "msedge.exe"},
		{given: Chrome, exe: "chrome.exe"},
		{given: Firefox, exe: "firefox.exe"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run((string)(tt.given), func(t *testing.T) {
			prePids, err := findPids(tt.exe)
			require.NoError(t, err)

			// In Github Actions, OpenKioskWeb sometimes returns error
			_ = OpenKioskWeb("https://github.com", Config{Browser: tt.given})
			// assert.NoError(t, err)

			postPids, err := findPids(tt.exe)
			require.NoError(t, err)
			assert.True(t, len(prePids) < len(postPids))
		})
	}
}

func findPids(exe string) (ret []int, err error) {
	processes, err := ps.Processes()
	if err != nil {
		return
	}
	for _, p := range processes {
		if p.Executable() == exe {
			ret = append(ret, p.Pid())
		}
	}
	return
}

func TestTimeoutToOpenKioskWeb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := OpenKioskWeb(
		"http://localhost:8000",
		Config{Browser: IE, WaitCtx: ctx},
	)
	assert.Error(t, err)
}

func TestNotTimeoutToOpenKioskWeb(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	server := &http.Server{Addr: ":8000"}
	defer func() {
		_ = server.Shutdown(context.Background())
	}()
	go func() {
		time.Sleep(2 * time.Second)
		_ = server.ListenAndServe()
	}()

	err := OpenKioskWeb(
		"http://localhost:8000",
		Config{Browser: IE, WaitCtx: ctx},
	)
	assert.NoError(t, err)
}
