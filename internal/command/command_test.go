package command

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestArgumentsAndVersion(t *testing.T) {
	for _, test := range []struct {
		args   []string
		status int
	}{
		{nil, 0}, {[]string{"--help"}, 0}, {[]string{"version"}, 0},
		{[]string{"unknown"}, 2}, {[]string{"serve", "--help"}, 0},
		{[]string{"serve", "--unknown"}, 2}, {[]string{"serve", "--listen="}, 2},
		{[]string{"serve", "extra"}, 2},
	} {
		var stdout, stderr bytes.Buffer
		if result := Run(context.Background(), test.args, &stdout, &stderr); result != test.status {
			t.Fatalf("%v: %d: %s", test.args, result, &stderr)
		}
		if len(test.args) == 1 && test.args[0] == "version" && stdout.String() != Version+"\n" {
			t.Fatalf("version = %q", stdout.String())
		}
	}
}

func TestCancelledServerShutsDownWithoutUpstream(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "index.html"), []byte("portal"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	if result := Run(ctx, []string{"serve", "--listen=127.0.0.1:0", "--assets=" + directory}, &stdout, &stderr); result != 0 {
		t.Fatalf("exit %d: %s", result, &stderr)
	}
}
