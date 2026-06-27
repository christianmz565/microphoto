package coordinator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type Effect struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

func (h *HTTPHandler) PreviewImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(h.maxUploadSize)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	effectsJSON := r.FormValue("effects")

	var effects []Effect
	if effectsJSON != "" {
		if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
			http.Error(w, "Invalid effects JSON", http.StatusBadRequest)
			return
		}
	}

	// Check if it's a video file
	isVideo := isVideoFile(header.Filename, data)

	var img image.Image
	if isVideo {
		img, err = extractFirstFrame(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to extract frame from video: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			http.Error(w, "Failed to decode image", http.StatusBadRequest)
			return
		}
	}

	for _, effect := range effects {
		img = applyEffect(img, effect.Type, effect.Params)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(buf.Bytes())
}

func isVideoFile(filename string, data []byte) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".flv", ".wmv", ".m4v":
		return true
	}
	// Check magic bytes for common video formats
	if len(data) > 12 {
		// MP4/MOV: ftyp at offset 4
		if string(data[4:8]) == "ftyp" {
			return true
		}
		// WebM/MKV: EBML header
		if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
			return true
		}
	}
	return false
}

func extractFirstFrame(videoData []byte) (image.Image, error) {
	tmpDir, err := os.MkdirTemp("", "preview-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpVideo := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(tmpVideo, videoData, 0644); err != nil {
		return nil, fmt.Errorf("write temp video: %w", err)
	}

	tmpFrame := filepath.Join(tmpDir, "frame.png")
	cmd := exec.Command("ffmpeg", "-i", tmpVideo, "-vframes", "1", "-f", "image2", tmpFrame)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg extract: %w: %s", err, stderr.String())
	}

	frameData, err := os.ReadFile(tmpFrame)
	if err != nil {
		return nil, fmt.Errorf("read frame: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(frameData))
	if err != nil {
		return nil, fmt.Errorf("decode frame: %w", err)
	}

	return img, nil
}

func applyEffect(img image.Image, effectType string, params map[string]string) image.Image {
	switch effectType {
	case "GRAYSCALE":
		return applyGrayscale(img)
	case "BLUR":
		radius := 1.0
		if v, err := strconv.ParseFloat(params["radius"], 64); err == nil {
			radius = v
		}
		return applyBoxBlur(img, radius)
	case "BRIGHTNESS":
		factor := 1.0
		if v, err := strconv.ParseFloat(params["factor"], 64); err == nil {
			factor = v
		}
		return applyBrightness(img, factor)
	case "RESIZE":
		w, _ := strconv.Atoi(params["width"])
		h, _ := strconv.Atoi(params["height"])
		if w > 0 && h > 0 {
			return applyResize(img, w, h)
		}
		return img
	default:
		return img
	}
}

func applyGrayscale(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			lum := uint16(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
			result.SetRGBA(x, y, color.RGBA{R: uint8(lum >> 8), G: uint8(lum >> 8), B: uint8(lum >> 8), A: uint8(a >> 8)})
		}
	}
	return result
}

func applyBrightness(img image.Image, factor float64) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			result.SetRGBA(x, y, color.RGBA{
				R: clamp(float64(r>>8) * factor),
				G: clamp(float64(g>>8) * factor),
				B: clamp(float64(b>>8) * factor),
				A: uint8(a >> 8),
			})
		}
	}
	return result
}

func applyBoxBlur(img image.Image, sigma float64) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	radius := int(math.Ceil(sigma * 1.5))
	if radius < 1 {
		radius = 1
	}

	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, img, bounds.Min, draw.Src)
	dst := image.NewRGBA(bounds)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var rSum, gSum, bSum, aSum float64
			var count float64
			for dx := -radius; dx <= radius; dx++ {
				xx := x + dx
				if xx >= 0 && xx < w {
					c := src.RGBAAt(xx, y)
					rSum += float64(c.R)
					gSum += float64(c.G)
					bSum += float64(c.B)
					aSum += float64(c.A)
					count++
				}
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(rSum / count),
				G: uint8(gSum / count),
				B: uint8(bSum / count),
				A: uint8(aSum / count),
			})
		}
	}

	result := image.NewRGBA(bounds)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var rSum, gSum, bSum, aSum float64
			var count float64
			for dy := -radius; dy <= radius; dy++ {
				yy := y + dy
				if yy >= 0 && yy < h {
					c := dst.RGBAAt(x, yy)
					rSum += float64(c.R)
					gSum += float64(c.G)
					bSum += float64(c.B)
					aSum += float64(c.A)
					count++
				}
			}
			result.SetRGBA(x, y, color.RGBA{
				R: uint8(rSum / count),
				G: uint8(gSum / count),
				B: uint8(bSum / count),
				A: uint8(aSum / count),
			})
		}
	}

	return result
}

func applyResize(img image.Image, targetW, targetH int) image.Image {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	result := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			srcX := bounds.Min.X + x*srcW/targetW
			srcY := bounds.Min.Y + y*srcH/targetH
			result.SetRGBA(x, y, img.At(srcX, srcY).(color.RGBA))
		}
	}
	return result
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
