// Package httpapi serves the portal and reads observations from Homehub.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const maxResponseBytes = 1 << 20

type observation struct {
	ID                int64     `json:"id"`
	ObservedAt        time.Time `json:"observed_at"`
	Address           string    `json:"address"`
	Model             string    `json:"model"`
	TemperatureC      *float64  `json:"temperature_c"`
	TemperatureF      float64   `json:"temperature_f"`
	HumidityPercent   *int      `json:"humidity_percent"`
	BatteryPercent    *int      `json:"battery_percent"`
	CO2PPM            *int      `json:"co2_ppm"`
	RSSIDBM           *int      `json:"rssi_dbm"`
	FahrenheitDisplay bool      `json:"fahrenheit_display"`
}

type deviceResponse struct {
	Devices []observation `json:"devices"`
}

func New(source string, web http.Handler, logger *slog.Logger) (http.Handler, error) {
	base, err := url.Parse(source)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") ||
		base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("--homehub-url must be an HTTP(S) base URL without credentials, query or fragment")
	}
	endpoint := base.JoinPath("v1", "devices").String()
	diskEndpoint := base.JoinPath("v1", "server", "disks").String()
	client := &http.Client{
		Timeout:       4 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/devices", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()
		body, err := readDevices(ctx, client, endpoint)
		if err != nil {
			logger.Warn("read observations", "error", err)
			w.Header().Set("Retry-After", "5")
			writeJSON(w, r, http.StatusServiceUnavailable, map[string]string{"error": "機器のデータを取得できませんでした。時間をおいて再読み込みしてください。"})
			return
		}
		writeJSON(w, r, http.StatusOK, body)
	})
	mux.HandleFunc("/api/devices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, r, http.StatusMethodNotAllowed, map[string]string{"error": "GET または HEAD を使用してください。"})
	})
	mux.HandleFunc("GET /api/disks", func(w http.ResponseWriter, r *http.Request) {
		body, err := readDisks(r.Context(), client, diskEndpoint)
		if err != nil {
			logger.Warn("read disk usage", "error", err)
			writeJSON(w, r, http.StatusServiceUnavailable, map[string]string{"error": "空き容量の結果を取得できませんでした。時間をおいて再読み込みしてください。"})
			return
		}
		writeJSON(w, r, http.StatusOK, body)
	})
	mux.HandleFunc("/api/disks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", "GET, HEAD")
		writeJSON(w, r, http.StatusMethodNotAllowed, map[string]string{"error": "GET または HEAD を使用してください。"})
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, r, http.StatusNotFound, map[string]string{"error": "API が見つかりません。"})
	})
	mux.Handle("/", web)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		mux.ServeHTTP(w, r)
	}), nil
}

func readDevices(ctx context.Context, client *http.Client, endpoint string) (deviceResponse, error) {
	data, err := readResponse(ctx, client, endpoint)
	if err != nil {
		return deviceResponse{}, err
	}
	var body deviceResponse
	if err := json.Unmarshal(data, &body); err != nil {
		return deviceResponse{}, err
	}
	if body.Devices == nil {
		return deviceResponse{}, errors.New("Homehub response is missing the devices array")
	}
	for _, device := range body.Devices {
		if device.Address == "" || device.ObservedAt.IsZero() {
			return deviceResponse{}, errors.New("Homehub observation is missing its address or timestamp")
		}
		if device.TemperatureC == nil || device.HumidityPercent == nil {
			return deviceResponse{}, errors.New("Homehub observation is missing its temperature or humidity")
		}
	}
	return body, nil
}

func readResponse(ctx context.Context, client *http.Client, endpoint string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("Homehub returned " + response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxResponseBytes {
		return nil, errors.New("Homehub response exceeds size limit")
	}
	return data, nil
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, value any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		json.NewEncoder(w).Encode(value)
	}
}
