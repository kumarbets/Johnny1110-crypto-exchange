package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "/app/fe-dist"
	fs := http.FileServer(http.Dir(dir))

	serveIndex := func(w http.ResponseWriter, r *http.Request) {
		// index.html must never be cached, so a rebuild's new hashed JS/CSS is
		// always picked up. (The hashed assets themselves are safe to cache.)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		full := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r) // real file (hashed js/css/img/favicon)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/js/") || strings.HasPrefix(r.URL.Path, "/css/") ||
			strings.HasPrefix(r.URL.Path, "/img/") || r.URL.Path == "/favicon.ico" {
			fs.ServeHTTP(w, r)
			return
		}
		serveIndex(w, r) // SPA fallback + root
	})
	log.Println("FE static server on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
