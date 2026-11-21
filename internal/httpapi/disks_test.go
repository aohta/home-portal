package httpapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const diskFixture = `{"collected_at":"2026-11-21T00:00:00Z","stale":true,"filesystems":[{"name":"Samba","path":"/srv/samba/shared","status":"ok","total_bytes":1000,"used_bytes":950,"available_bytes":0}]}`

func TestDiskProxyPreservesStaleAndZeroCapacity(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prefix/v1/server/disks" || r.Method != "GET" {
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Cookie") != "" {
			t.Error("forwarded browser cookie")
		}
		io.WriteString(w, diskFixture)
	}))
	defer upstream.Close()
	handler := portal(t, upstream.URL+"/prefix")
	request := httptest.NewRequest("GET", "/api/disks", nil)
	request.Header.Set("Cookie", "session=private")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"stale":true`) || !strings.Contains(response.Body.String(), `"available_bytes":0`) {
		t.Fatalf("response: %d %s", response.Code, response.Body)
	}
	for _, method := range []string{"HEAD", "POST"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(method, "/api/disks", nil))
		if method == "HEAD" && (response.Code != 200 || response.Body.Len() != 0) {
			t.Fatal("invalid HEAD response")
		}
		if method == "POST" && response.Code != 405 {
			t.Fatal("accepted write request")
		}
	}
}

func TestInvalidDiskResponsesAreUnavailable(t *testing.T) {
	for _, body := range []string{`{}`, `{"collected_at":"2026-11-21T00:00:00Z","stale":false,"filesystems":[]}`,
		`{"collected_at":"2026-11-21T00:00:00Z","stale":false,"filesystems":[{"name":"Samba","path":"/share","status":"ok"}]}`} {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) }))
		response := httptest.NewRecorder()
		portal(t, upstream.URL).ServeHTTP(response, httptest.NewRequest("GET", "/api/disks", nil))
		upstream.Close()
		if response.Code != 503 {
			t.Fatalf("accepted %s", body)
		}
	}
}
