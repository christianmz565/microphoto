package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"maps"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
)

// Effect represents an image processing effect with its parameters.
type Effect struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

// PreviewImage handles the multipart form upload of an image for preview with effects applied.
func (h *HTTPHandler) PreviewImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}
		} else {
			err = r.ParseMultipartForm(h.maxUploadSize)
			if err != nil && !errors.Is(err, http.ErrNotMultipart) {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			} else if errors.Is(err, http.ErrNotMultipart) {
				_ = r.ParseForm()
			}
		}
	}

	previewID := r.FormValue("preview_id")
	effectsJSON := r.FormValue("effects")

	var effects []Effect
	if effectsJSON != "" {
		if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
			http.Error(w, "Invalid effects JSON", http.StatusBadRequest)
			return
		}
	}

	var (
		cutData  []byte
		isVideo  bool
		filename string
	)

	if previewID != "" {
		val, ok := h.previewCache.Load(previewID)
		if !ok {
			http.Error(w, "Preview session expired or not found", http.StatusNotFound)
			return
		}

		cp := val.(cachedPreview)
		cutData = cp.data
		isVideo = true
		filename = "preview.mp4"
	} else {
		// Normal upload
		file, header, err := r.FormFile("image")
		if err != nil {
			http.Error(w, "Image file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename = header.Filename
		isVideo = isVideoFile(filename, nil)

		if isVideo {
			tmpInput, err := os.CreateTemp("", "preview-upload-*.mp4")
			if err != nil {
				http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
				return
			}

			tmpInputPath := tmpInput.Name()
			defer os.Remove(tmpInputPath)
			defer tmpInput.Close()

			if _, err := io.Copy(tmpInput, file); err != nil {
				http.Error(w, "Failed to save upload", http.StatusInternalServerError)
				return
			}

			tmpInput.Close()

			cutData, err = cutVideoPreviewFile(r.Context(), tmpInputPath, 2)
			if err != nil {
				log.Printf("Failed to cut preview video: %v", err)
				http.Error(w, "Failed to generate video preview", http.StatusInternalServerError)
				return
			}

			previewID = "preview-vid-" + uuid.New().String()
			h.previewCache.Store(previewID, cachedPreview{
				data:      cutData,
				createdAt: time.Now(),
			})
		} else {
			data, err := io.ReadAll(file)
			if err != nil {
				http.Error(w, "Failed to read file", http.StatusInternalServerError)
				return
			}

			cutData = data
		}
	}

	if isVideo {
		taskID := "preview-vid-" + uuid.New().String()

		targetType := jobs.JobType_JOB_TYPE_GRAYSCALE
		if len(effects) > 0 {
			targetType = parseJobType(effects[0].Type)
		}

		params := make(map[string]string)
		if len(effects) > 0 {
			maps.Copy(params, map[string]string(effects[0].Params))
		}

		if effectsJSON != "" {
			params["effects"] = effectsJSON
		}

		// Start distributed video processing
		err = h.orchestrator.ProcessVideo(r.Context(), taskID, bytes.NewReader(cutData), filename, targetType, int64(len(cutData)), params)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start distributed video preview: %v", err), http.StatusInternalServerError)
			return
		}

		// Subscribe to progress events to wait for completion
		pubsub, ch := h.orchestrator.redis.SubscribeProgress(r.Context(), taskID)
		defer pubsub.Close()

		// Wait for completion (or timeout/failure)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		completed := false
		for !completed {
			select {
			case <-ctx.Done():
				http.Error(w, "Video preview generation timed out", http.StatusGatewayTimeout)
				return
			case msg, ok := <-ch:
				if !ok {
					http.Error(w, "Progress channel closed", http.StatusInternalServerError)
					return
				}

				var event model.ProgressPayload
				if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
					switch event.Status {
					case "JOB_COMPLETED":
						completed = true
					case "JOB_FAILED":
						http.Error(w, "Distributed preview failed: "+event.Message, http.StatusInternalServerError)
						return
					}
				}
			}
		}

		reader, err := h.orchestrator.DownloadVideoResult(r.Context(), taskID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to download video result: %v", err), http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Type", "video/mp4")

		if previewID != "" {
			w.Header().Set("X-Preview-ID", previewID)
			w.Header().Set("Access-Control-Expose-Headers", "X-Preview-ID")
		}

		io.Copy(w, reader)

		return
	}

	img, _, err := image.Decode(bytes.NewReader(cutData))
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusBadRequest)
		return
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

	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func isVideoFile(filename string, data []byte) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".flv", ".wmv", ".m4v":
		return true
	}

	if len(data) > 12 {
		if string(data[4:8]) == "ftyp" {
			return true
		}

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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpVideo := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(tmpVideo, videoData, 0o644); err != nil {
		return nil, fmt.Errorf("write temp video: %w", err)
	}

	tmpFrame := filepath.Join(tmpDir, "frame.png")
	cmd := exec.CommandContext(context.Background(), "ffmpeg", "-i", tmpVideo, "-vframes", "1", "-f", "image2", tmpFrame)

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

func cutVideoPreviewFile(ctx context.Context, inputPath string, durationSec int) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "preview-cut-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpOutput := filepath.Join(tmpDir, "output.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-ss", "0", "-t", strconv.Itoa(durationSec), "-i", inputPath, "-c", "copy", "-map", "0", "-avoid_negative_ts", "1", tmpOutput)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg cut: %w: %s", err, stderr.String())
	}

	outputData, err := os.ReadFile(tmpOutput)
	if err != nil {
		return nil, fmt.Errorf("read cut video: %w", err)
	}

	return outputData, nil
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
		if scaleStr, ok := params["scale"]; ok {
			scale, err := strconv.ParseFloat(scaleStr, 64)
			if err == nil && scale > 0 {
				bounds := img.Bounds()
				targetW := int(float64(bounds.Dx()) * scale)
				targetH := int(float64(bounds.Dy()) * scale)
				if targetW > 0 && targetH > 0 {
					return applyResize(img, targetW, targetH)
				}
			}
		}

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

	radius := max(int(math.Ceil(sigma*1.5)), 1)

	src := image.NewRGBA(bounds)
	draw.Draw(src, bounds, img, bounds.Min, draw.Src)
	dst := image.NewRGBA(bounds)

	blurPass(dst, src, w, h, radius, true)

	result := image.NewRGBA(bounds)

	blurPass(result, dst, w, h, radius, false)

	return result
}

func blurPass(dst, src *image.RGBA, w, h, radius int, horizontal bool) {
	for y := range h {
		for x := range w {
			var (
				rSum, gSum, bSum, aSum float64
				count                  float64
			)

			for d := -radius; d <= radius; d++ {
				var px, py int
				if horizontal {
					px, py = x+d, y
				} else {
					px, py = x, y+d
				}

				if px >= 0 && px < w && py >= 0 && py < h {
					c := src.RGBAAt(px, py)
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
}

func applyResize(img image.Image, targetW, targetH int) image.Image {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	result := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	for y := range targetH {
		for x := range targetW {
			srcX := bounds.Min.X + x*srcW/targetW
			srcY := bounds.Min.Y + y*srcH/targetH

			c, ok := color.NRGBAModel.Convert(img.At(srcX, srcY)).(*color.NRGBA)
			if !ok {
				continue
			}

			result.SetRGBA(x, y, color.RGBA{
				R: c.R,
				G: c.G,
				B: c.B,
				A: c.A,
			})
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
