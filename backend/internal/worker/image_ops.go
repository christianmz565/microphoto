package worker

import (
	"image"
	"image/color"
	"math"
)

// ApplyGrayscale converts an image to grayscale.
func ApplyGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	grayImg := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayImg.Set(x, y, img.At(x, y))
		}
	}
	return grayImg
}

// ApplyBrightness adjusts the brightness of an image by a given factor.
func ApplyBrightness(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()

			nr := uint8(math.Min(255, float64(r>>8)*factor))
			ng := uint8(math.Min(255, float64(g>>8)*factor))
			nb := uint8(math.Min(255, float64(b>>8)*factor))
			na := uint8(a >> 8)

			newImg.Set(x, y, color.RGBA{nr, ng, nb, na})
		}
	}
	return newImg
}

// ApplyBlur applies a simple box blur to the image with a given radius.
func ApplyBlur(img image.Image, radius int) image.Image {
	if radius <= 0 {
		return img
	}
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum, aSum uint32
			var count uint32

			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
						r, g, b, a := img.At(nx, ny).RGBA()
						rSum += r >> 8
						gSum += g >> 8
						bSum += b >> 8
						aSum += a >> 8
						count++
					}
				}
			}

			newImg.Set(x, y, color.RGBA{
				uint8(rSum / count),
				uint8(gSum / count),
				uint8(bSum / count),
				uint8(aSum / count),
			})
		}
	}
	return newImg
}

// ApplyResize resizes an image to the target dimensions using Nearest-Neighbor interpolation.
func ApplyResize(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	if targetWidth <= 0 || targetHeight <= 0 {
		return img
	}

	newImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	for y := range targetHeight {
		for x := 0; x < targetWidth; x++ {

			srcX := int(float64(x) * float64(originalWidth) / float64(targetWidth))
			srcY := int(float64(y) * float64(originalHeight) / float64(targetHeight))

			if srcX >= originalWidth {
				srcX = originalWidth - 1
			}
			if srcY >= originalHeight {
				srcY = originalHeight - 1
			}

			newImg.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return newImg
}
