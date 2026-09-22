package imageutil_test

import (
	"strings"
	"testing"

	"github.com/KarpovYuri/caraudio-backend/internal/catalog/infrastructure/imageutil"
)

func TestSanitizeSVGRemovesScript(t *testing.T) {
	raw := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10">
  <script>alert(1)</script>
  <circle cx="5" cy="5" r="4" fill="red"/>
</svg>`)

	clean, err := imageutil.SanitizeSVG(raw, 1<<20)
	if err != nil {
		t.Fatalf("SanitizeSVG: %v", err)
	}
	out := string(clean)
	if strings.Contains(strings.ToLower(out), "<script") {
		t.Fatalf("script tag was not removed: %s", out)
	}
	if !strings.Contains(out, "circle") {
		t.Fatalf("expected circle to remain: %s", out)
	}
}

func TestIsSVG(t *testing.T) {
	if !imageutil.IsSVG("logo.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)) {
		t.Fatal("expected svg by extension/content")
	}
	if imageutil.IsSVG("logo.png", []byte{0x89, 0x50, 0x4e, 0x47}) {
		t.Fatal("png should not be detected as svg")
	}
}
