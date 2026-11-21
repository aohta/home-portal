package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type diskFilesystem struct {
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	Status         string  `json:"status"`
	TotalBytes     *uint64 `json:"total_bytes"`
	UsedBytes      *uint64 `json:"used_bytes"`
	AvailableBytes *uint64 `json:"available_bytes"`
	Error          string  `json:"error,omitempty"`
}

type diskSnapshot struct {
	CollectedAt time.Time        `json:"collected_at"`
	Stale       *bool            `json:"stale"`
	Filesystems []diskFilesystem `json:"filesystems"`
}

func readDisks(ctx context.Context, client *http.Client, endpoint string) (diskSnapshot, error) {
	data, err := readResponse(ctx, client, endpoint)
	if err != nil {
		return diskSnapshot{}, err
	}
	var body diskSnapshot
	if err := json.Unmarshal(data, &body); err != nil {
		return diskSnapshot{}, err
	}
	if body.CollectedAt.IsZero() || body.Stale == nil || len(body.Filesystems) < 1 || len(body.Filesystems) > 16 {
		return diskSnapshot{}, errors.New("invalid disk snapshot")
	}
	for _, row := range body.Filesystems {
		if row.Name == "" || row.Path == "" || (row.Status != "ok" && row.Status != "error") {
			return diskSnapshot{}, errors.New("invalid filesystem identity or status")
		}
		if row.Status == "ok" && (row.TotalBytes == nil || row.UsedBytes == nil || row.AvailableBytes == nil ||
			*row.TotalBytes == 0 || *row.UsedBytes > *row.TotalBytes || *row.AvailableBytes > *row.TotalBytes-*row.UsedBytes) {
			return diskSnapshot{}, errors.New("invalid filesystem capacity")
		}
	}
	return body, nil
}
