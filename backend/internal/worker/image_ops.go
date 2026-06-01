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

// ApplyBlur applies a simple 3x3 box blur to the image.
func ApplyBlur(img image.Image) image.Image {
	bounds := img.Bounds()
	newImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var rSum, gSum, bSum, aSum uint32
			var count uint32

			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
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
