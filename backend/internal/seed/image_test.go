package seed

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderImageIsDeterministicPNG(t *testing.T) {
	first, err := RenderImage(12, 1, 3, 42)
	if err != nil {
		t.Fatal(err)
	}
	second, err := RenderImage(12, 1, 3, 42)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("same inputs produced different images")
	}
	decoded, err := png.Decode(bytes.NewReader(first))
	if err != nil {
		t.Fatalf("decode generated PNG: %v", err)
	}
	if decoded.Bounds().Dx() != SeedImageWidth || decoded.Bounds().Dy() != SeedImageHeight {
		t.Fatalf("unexpected dimensions: %v", decoded.Bounds())
	}
}
