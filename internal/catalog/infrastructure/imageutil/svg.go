package imageutil

import (
	"bytes"
	"fmt"
	"strings"

	svgsanitize "go.privatebychoice.com/pbcsvgsanitize"
)

const DefaultMaxSVGBytes = 1 << 20 // 1 MiB after sanitize budget for logos

// IsSVG reports whether raw looks like an SVG document.
func IsSVG(filename string, raw []byte) bool {
	ext := strings.ToLower(strings.TrimSpace(filename))
	if strings.HasSuffix(ext, ".svg") {
		return true
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return false
	}
	lower := bytes.ToLower(trimmed)
	if bytes.HasPrefix(lower, []byte("<svg")) {
		return true
	}
	if bytes.HasPrefix(lower, []byte("<?xml")) && bytes.Contains(lower, []byte("<svg")) {
		return true
	}
	return false
}

// SanitizeSVG reduces an untrusted SVG to a safe presentational subset.
func SanitizeSVG(raw []byte, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxSVGBytes
	}
	if len(raw) > maxBytes {
		return nil, fmt.Errorf("svg file is too large")
	}

	clean, _, err := svgsanitize.Sanitize(raw,
		svgsanitize.WithMaxBytes(maxBytes),
		svgsanitize.WithMaxDepth(64),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid or unsafe svg: %w", err)
	}
	if len(bytes.TrimSpace(clean)) == 0 {
		return nil, fmt.Errorf("svg is empty after sanitization")
	}
	return clean, nil
}
