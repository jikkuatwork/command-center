package hosting

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// ServeVFS serves files from the Virtual File System
func ServeVFS(w http.ResponseWriter, r *http.Request, siteID string) {
	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	// Default to index.html for root or directories
	if path == "/" || strings.HasSuffix(path, "/") {
		path += "index.html"
	}

	// Clean path
	path = filepath.Clean(path)
	// Ensure consistent forward slashes
	path = filepath.ToSlash(path)
	// Remove leading slash for DB lookup if stored without it
	// In deploy.go we used filepath.Clean/ToSlash, which usually removes leading slash for relative paths?
	// zip files usually don't have leading slash.
	// Let's ensure we strip leading slash.
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// 1. Try exact match
	file, err := fs.ReadFile(siteID, path)
	if err != nil {
		// 2. If not found, and it looks like a directory (no extension), try appending index.html
		if filepath.Ext(path) == "" {
			idxPath := filepath.Join(path, "index.html")
			idxPath = filepath.ToSlash(idxPath)
			file, err = fs.ReadFile(siteID, idxPath)
		}
		
		// 3. If still not found, 404
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	defer file.Content.Close()

	// ETag Caching
	w.Header().Set("ETag", fmt.Sprintf(`"%s"`, file.Hash))
	if match := r.Header.Get("If-None-Match"); match != "" {
		if strings.Contains(match, file.Hash) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Content Type
	contentType := file.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}
	w.Header().Set("Content-Type", contentType)
	
	// Content Length
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))

	// Serve content
	if _, err := io.Copy(w, file.Content); err != nil {
		// Log error?
	}
}
