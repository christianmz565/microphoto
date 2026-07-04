package coordinator

import (
	"image"
	"image/color"
	"testing"
)

func createPreviewTestImage(width, height int) image.Image {
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

	return img
}

func BenchmarkApplyGrayscale(b *testing.B) {
	img := createPreviewTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_ = applyGrayscale(img)
	}
}

func BenchmarkApplyBrightness(b *testing.B) {
	img := createPreviewTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_ = applyBrightness(img, 1.5)
	}
}

func BenchmarkApplyBoxBlur(b *testing.B) {
	img := createPreviewTestImage(256, 256)

	b.ReportAllocs()

	for b.Loop() {
		_ = applyBoxBlur(img, 2.0)
	}
}

func BenchmarkApplyResize(b *testing.B) {
	img := createPreviewTestImage(512, 512)

	b.ReportAllocs()

	for b.Loop() {
		_ = applyResize(img, 256, 256)
	}
}

func BenchmarkClamp(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		_ = clamp(300.0)
	}
}
