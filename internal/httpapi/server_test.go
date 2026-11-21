package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const fixture = `{"devices":[{"id":1,"observed_at":"2026-11-18T00:00:00Z","address":"E7:50:54:EF:DF:F8","temperature_c":0,"humidity_percent":0,"battery_percent":0,"co2_ppm":null}]}`

func portal(t *testing.T, source string) http.Handler {
	t.Helper()
	handler, err := New(source, http.NotFoundHandler(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func TestReadOnlyProxyPreservesZeroAndNullWithoutForwardingCredentials(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/prefix/v1/devices" || r.URL.RawQuery != "" {
			t.Errorf("unexpected upstream request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
			t.Error("forwarded browser credentials")
		}
		io.WriteString(w, fixture)
	}))
	defer upstream.Close()
	handler := portal(t, upstream.URL+"/prefix/")
	request := httptest.NewRequest("GET", "/api/devices", nil)
	request.Header.Set("Cookie", "session=private")
	request.Header.Set("Authorization", "Bearer private")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 200 || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response: %d %s", response.Code, response.Body)
	}
	var body deviceResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	row := body.Devices[0]
	if row.TemperatureC == nil || *row.TemperatureC != 0 || row.HumidityPercent == nil || *row.HumidityPercent != 0 ||
		row.BatteryPercent == nil || *row.BatteryPercent != 0 || row.CO2PPM != nil {
		t.Fatalf("zero/null changed: %+v", row)
	}
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("HEAD", "/api/devices", nil))
	if response.Code != 200 || response.Body.Len() != 0 {
		t.Fatalf("HEAD: %d %s", response.Code, response.Body)
	}
}

func TestUpstreamFailuresDoNotBecomeSuccessfulReadings(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"unavailable", 503, "private database path"},
		{"redirect", 302, ""},
		{"invalid JSON", 200, "<html>login</html>"},
		{"missing devices", 200, `{}`},
		{"null devices", 200, `{"devices":null}`},
		{"missing timestamp", 200, `{"devices":[{"address":"E7:50:54:EF:DF:F8"}]}`},
		{"missing temperature", 200, strings.Replace(fixture, `"temperature_c":0,`, "", 1)},
		{"null temperature", 200, strings.Replace(fixture, `"temperature_c":0`, `"temperature_c":null`, 1)},
		{"missing humidity", 200, strings.Replace(fixture, `"humidity_percent":0,`, "", 1)},
		{"null humidity", 200, strings.Replace(fixture, `"humidity_percent":0`, `"humidity_percent":null`, 1)},
		{"oversized", 200, strings.Repeat(" ", maxResponseBytes+1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "/must-not-follow")
				w.WriteHeader(test.status)
				io.WriteString(w, test.body)
			}))
			defer upstream.Close()
			response := httptest.NewRecorder()
			portal(t, upstream.URL).ServeHTTP(response, httptest.NewRequest("GET", "/api/devices", nil))
			if response.Code != 503 || response.Header().Get("Retry-After") != "5" || !json.Valid(response.Body.Bytes()) {
				t.Fatalf("response: %d %s", response.Code, response.Body)
			}
			if strings.Contains(response.Body.String(), "private database path") {
				t.Fatal("exposed upstream error")
			}
		})
	}
}

func TestUnavailableUpstreamDoesNotBlockPortalOrHealth(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	upstream.Close()
	handler := portal(t, upstream.URL)
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/healthz", 200}, {"GET", "/api/devices", 503},
		{"POST", "/api/devices", 405}, {"GET", "/api/missing", 404},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.status {
			t.Fatalf("%s %s: %d", test.method, test.path, response.Code)
		}
	}
}

func TestInvalidSourceConfiguration(t *testing.T) {
	for _, source := range []string{"", "/relative", "file:///private", "http://user:password@localhost", "http://localhost?query=1", "http://localhost/#fragment"} {
		if _, err := New(source, http.NotFoundHandler(), slog.Default()); err == nil {
			t.Errorf("accepted %q", source)
		}
	}
}
