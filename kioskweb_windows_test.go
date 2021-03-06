package kioskweb

import (
	"context"
	"net/http"
	"os/exec"
	"strconv"
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

			err = Open("https://github.com", Config{Browser: tt.given})
			assert.NoError(t, err)

			postPids, err := findPids(tt.exe)
			require.NoError(t, err)
			assert.True(t, len(prePids) < len(postPids))
		L:
			for _, post := range postPids {
				for _, pre := range prePids {
					if pre == post {
						continue L
					}
				}
				_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(post)).Run()
			}
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

func kill(exe string) error {
	ret, err := findPids(exe)
	if err != nil {
		return err
	}
	for _, pid := range ret {
		_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(pid)).Run()
	}
	return nil
}

func TestTimeoutToOpenKioskWeb(t *testing.T) {
	defer kill("iexplore.exe")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := Open(
		"http://localhost:8000",
		Config{Browser: IE, WaitCtx: ctx},
	)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestNotTimeoutToOpenKioskWeb(t *testing.T) {
	defer kill("iexplore.exe")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	server := &http.Server{Addr: ":8000"}
	defer func() {
		_ = server.Shutdown(context.Background())
	}()
	go func() {
		time.Sleep(2 * time.Second)
		_ = server.ListenAndServe()
	}()

	err := Open(
		"http://localhost:8000",
		Config{Browser: IE, WaitCtx: ctx},
	)
	assert.NoError(t, err)
}
