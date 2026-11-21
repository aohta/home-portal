package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
)

func Handler(directory string) (http.Handler, error) {
	files, err := embeddedFiles()
	if err != nil {
		return nil, err
	}
	if directory != "" {
		files = os.DirFS(directory)
	}
	if files == nil {
		return nil, errors.New("frontend not built: use make build, or --assets apps/web/dist for development")
	}
	return fromFS(files)
}

func fromFS(files fs.FS) (http.Handler, error) {
	if _, err := fs.ReadFile(files, "index.html"); err != nil {
		return nil, err
	}
	static := http.FileServerFS(files)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		static.ServeHTTP(w, r)
	}), nil
}
