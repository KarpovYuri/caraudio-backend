package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp"
)

const DefaultMaxLogoSide = 512

// ProcessLogo decodes an uploaded image, fits it into maxSide×maxSide and encodes lossless WebP.
func ProcessLogo(r io.Reader, maxSide int) ([]byte, error) {
	if maxSide <= 0 {
		maxSide = DefaultMaxLogoSide
	}

	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("invalid image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image dimensions")
	}

	if w > maxSide || h > maxSide {
		img = imaging.Fit(img, maxSide, maxSide, imaging.Lanczos)
	}

	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, &nativewebp.Options{
		CompressionLevel: nativewebp.DefaultCompression,
	}); err != nil {
		return nil, fmt.Errorf("failed to encode webp: %w", err)
	}
	return buf.Bytes(), nil
}
