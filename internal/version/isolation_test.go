package version

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/marcus/sidecar/internal/hostexec"
	"github.com/marcus/sidecar/internal/testutil"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestMain(m *testing.M) {
	hostexec.LookPath = func(string) (string, error) { return "", exec.ErrNotFound }
	hostexec.Command = func(name string, args ...string) *exec.Cmd {
		return &exec.Cmd{Args: append([]string{name}, args...), Err: exec.ErrNotFound}
	}
	http.DefaultTransport = fixtureTransport(func(*http.Request) (*http.Response, error) { return nil, errors.New("fixture network unavailable") })
	os.Exit(testutil.Run(m))
}

func releaseResponse(t *testing.T, status int, body string) *int {
	t.Helper()
	calls := 0
	previous := http.DefaultTransport
	http.DefaultTransport = fixtureTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodGet || r.URL.String() != "https://api.github.com/repos/marcus/sidecar/releases/latest" {
			t.Fatalf("unexpected release request: %s %s", r.Method, r.URL)
		}
		return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = previous })
	return &calls
}
