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
