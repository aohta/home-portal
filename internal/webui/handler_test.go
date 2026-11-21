package webui

import (
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestAssetsAndMissingFiles(t *testing.T) {
	handler, err := fromFS(fstest.MapFS{
		"index.html":    {Data: []byte("<html>ホームポータル</html>")},
		"assets/app.js": {Data: []byte("console.log('portal')")},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{"GET", "/", 200}, {"GET", "/assets/app.js", 200},
		{"GET", "/assets/missing.js", 404}, {"GET", "/missing", 404},
		{"HEAD", "/", 200}, {"POST", "/", 405},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.status || (test.method == "HEAD" && response.Body.Len() != 0) {
			t.Fatalf("%+v: %d %s", test, response.Code, response.Body)
		}
	}
}
