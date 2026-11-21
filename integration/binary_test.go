//go:build integration

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestStandaloneBinary(t *testing.T) {
	binary := os.Getenv("HOME_PORTAL_TEST_BINARY")
	if binary == "" {
		t.Fatal("run make integration")
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/devices" {
			http.NotFound(w, r)
			return
		}
		io.WriteString(w, `{"devices":[]}`)
	}))
	defer upstream.Close()
	cmd := exec.Command(binary, "serve", "--listen=127.0.0.1:0", "--homehub-url="+upstream.URL)
	cmd.Dir = t.TempDir()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	wait := make(chan error, 1)
	go func() { wait <- cmd.Wait() }()
	t.Cleanup(func() {
		cmd.Process.Signal(os.Interrupt)
		select {
		case err := <-wait:
			if err != nil {
				t.Errorf("shutdown: %v", err)
			}
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
			<-wait
			t.Error("binary did not shut down")
		}
	})
	address := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			var entry struct {
				Address string `json:"address"`
			}
			if json.Unmarshal(scanner.Bytes(), &entry) == nil && entry.Address != "" {
				address <- entry.Address
				return
			}
		}
		address <- ""
	}()
	var base string
	select {
	case value := <-address:
		if value == "" {
			t.Fatal("binary did not report a listener")
		}
		base = "http://" + value
	case <-time.After(10 * time.Second):
		t.Fatal("binary startup timed out")
	}
	get := func(path string, expected int) string {
		t.Helper()
		client := &http.Client{Timeout: 5 * time.Second}
		response, err := client.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != expected {
			t.Fatal(fmt.Sprintf("%s: %d %s", path, response.StatusCode, body))
		}
		return string(body)
	}
	html := get("/", 200)
	if !strings.Contains(html, "ホームポータル") {
		t.Fatal("missing embedded frontend")
	}
	asset := regexp.MustCompile(`src="([^\"]+\.js)"`).FindStringSubmatch(html)
	if len(asset) != 2 {
		t.Fatal("missing frontend entry script")
	}
	get("/"+strings.TrimPrefix(asset[1], "./"), 200)
	get("/assets/does-not-exist.js", 404)
	if !strings.Contains(get("/api/devices", 200), `"devices":[]`) {
		t.Fatal("unexpected observations")
	}
	get("/healthz", 200)
}
