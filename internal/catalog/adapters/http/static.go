package httpadapter

import (
	"net/http"
	"strings"
)

// SecureStatic wraps a file server and adds hardening headers for SVG responses.
func SecureStatic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(strings.ToLower(r.URL.Path), ".svg") {
			w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
			// Isolate SVG if opened as a document; <img src> still works for display.
			w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
			w.Header().Set("X-Content-Type-Options", "nosniff")
		}
		next.ServeHTTP(w, r)
	})
}
