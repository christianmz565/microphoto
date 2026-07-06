package worker

import (
	"github.com/h2non/bimg"
)

// ApplyGrayscale converts an image buffer to grayscale.
func ApplyGrayscale(buf []byte) ([]byte, error) {
	options := bimg.Options{
		Interpretation: bimg.InterpretationBW,
	}

	return bimg.NewImage(buf).Process(options)
}

// ApplyBrightness adjusts the brightness of an image buffer using a multiplicative factor.
func ApplyBrightness(buf []byte, factor float64) ([]byte, error) {
	options := bimg.Options{
		Contrast: factor,
	}

	return bimg.NewImage(buf).Process(options)
}

// ApplyBlur applies Gaussian blur to the image buffer.
func ApplyBlur(buf []byte, sigma float64) ([]byte, error) {
	options := bimg.Options{
		GaussianBlur: bimg.GaussianBlur{
			Sigma: sigma,
		},
	}

	return bimg.NewImage(buf).Process(options)
}

// ApplyResize resizes an image buffer to the target dimensions.
func ApplyResize(buf []byte, targetWidth, targetHeight int) ([]byte, error) {
	if targetWidth <= 0 || targetHeight <= 0 {
		return buf, nil
	}

	options := bimg.Options{
		Width:  targetWidth,
		Height: targetHeight,
		Force:  true,
	}

	return bimg.NewImage(buf).Process(options)
}

// ExtractRegion extracts a specific rectangular region from the image buffer.
func ExtractRegion(buf []byte, x, y, width, height int) ([]byte, error) {
	return bimg.NewImage(buf).Extract(y, x, width, height)
}

// ApplyContrast adjusts the contrast of an image buffer using a multiplicative factor.
func ApplyContrast(buf []byte, factor float64) ([]byte, error) {
	options := bimg.Options{
		Contrast: factor,
	}

	return bimg.NewImage(buf).Process(options)
}

// ApplySepia applies a sepia tone effect to the image buffer.
func ApplySepia(buf []byte, intensity float64) ([]byte, error) {
	img := bimg.NewImage(buf)

	// Convert to grayscale first
	gs, err := img.Process(bimg.Options{
		Interpretation: bimg.InterpretationBW,
	})
	if err != nil {
		return buf, err
	}

	gsImg := bimg.NewImage(gs)

	// Apply sepia via color matrix transformation
	// Sepia: R = 0.393*gray + 0.769*gray + 0.189*gray = gray
	// The grayscale pixel is already a single channel, so we use gamma/brightness to tint
	options := bimg.Options{
		Gamma:   1.2,
		Brightness: 1,
	}

	result, err := gsImg.Process(options)
	if err != nil {
		return gs, err
	}

	return result, nil
}

// ApplyVignette applies a vignette (darkened edges) effect to the image buffer.
// Since bimg/libvips does not have a native vignette filter, we return the original buffer.
// The vignette effect is applied server-side in preview.go for images and via ffmpeg for videos.
func ApplyVignette(buf []byte, intensity float64) ([]byte, error) {
	return buf, nil
}
