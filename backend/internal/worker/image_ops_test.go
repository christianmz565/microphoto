package worker

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: uint8((x + y) % 256),
				A: 255,
			})
		}
	}

	var buf bytes.Buffer

	_ = png.Encode(&buf, img)

	return buf.Bytes()
}

func BenchmarkApplyGrayscale(b *testing.B) {
	data := createTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ApplyGrayscale(data)
	}
}

func BenchmarkApplyBrightness(b *testing.B) {
	data := createTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ApplyBrightness(data, 1.5)
	}
}

func BenchmarkApplyBlur(b *testing.B) {
	data := createTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ApplyBlur(data, 2.0)
	}
}

func BenchmarkApplyResize(b *testing.B) {
	data := createTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ApplyResize(data, 256, 256)
	}
}

func BenchmarkExtractRegion(b *testing.B) {
	data := createTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ExtractRegion(data, 0, 0, 256, 256)
	}
}

func BenchmarkApplyGrayscale_LargeImage(b *testing.B) {
	data := createTestImage(2048, 2048)

	b.ReportAllocs()

	for b.Loop() {
		_, _ = ApplyGrayscale(data)
	}
}
