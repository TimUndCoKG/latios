// In handler/frontend.go
package handler

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
)

// RegisterFrontendHandlers registers all routes for serving the SPA.
func RegisterFrontendHandlers(router *http.ServeMux, content embed.FS) error {
	log.Println("[ROUTER] Setting up frontend SPA routes...")

	// 1. Get the embedded 'dist' folder
	staticFiles, err := fs.Sub(content, "latios-frontend/dist")
	if err != nil {
		return fmt.Errorf("failed to get embedded filesystem: %w", err)
	}

	// 2. Static Asset Handler for frontend
	// Serves files from /latios/assets/ (e.g., /latios/assets/index-DVB8kM4A.css)
	log.Println("[ROUTER] Setting up static asset handler for /latios/assets/")
	assetServer := http.FileServer(http.FS(staticFiles))
	router.Handle(
		"/latios/assets/",
		http.StripPrefix("/latios/", assetServer),
	)

	// 3. SPA Fallback Handler
	// Serves index.html for any path under /latios/ that isn't an asset or API route
	log.Println("[ROUTER] Setting up SPA fallback handler for /latios/")
	router.HandleFunc("/latios/", func(w http.ResponseWriter, r *http.Request) {
		// Open index.html
		f, err := staticFiles.Open("index.html")
		if err != nil {
			log.Printf("[SPA] ERROR: Could not open index.html: %v", err)
			http.Error(w, "SPA not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			log.Printf("[SPA] ERROR: Could not stat index.html: %v", err)
			http.Error(w, "SPA stat error", http.StatusInternalServerError)
			return
		}

		// Serve the index.html file
		log.Printf("[SPA] Serving index.html for request: %s", r.URL.Path)
		http.ServeContent(w, r, "index.html", fi.ModTime(), f.(io.ReadSeeker))
	})

	return nil
}
