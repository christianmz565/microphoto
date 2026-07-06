package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type pipelineEffect struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

// buildFFmpegFilterChain converts a JSON effects array into an ffmpeg -vf filter chain string.
func buildFFmpegFilterChain(effectsJSON string) (string, error) {
	if effectsJSON == "" {
		return "", nil
	}

	var effects []pipelineEffect
	if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
		return "", fmt.Errorf("unmarshal effects: %w", err)
	}

	var filters []string

	for _, effect := range effects {
		switch effect.Type {
		case "GRAYSCALE":
			filters = append(filters, "hue=s=0")
		case "BLUR":
			radius := 1.0
			if r, err := strconv.ParseFloat(effect.Params["radius"], 64); err == nil {
				radius = r
			}
			filters = append(filters, fmt.Sprintf("gblur=sigma=%.2f", radius))
		case "BRIGHTNESS":
			factor := 1.0
			if f, err := strconv.ParseFloat(effect.Params["factor"], 64); err == nil {
				factor = f
			}
			eqBrightness := (factor - 1.0)

			if eqBrightness < -1.0 {
				eqBrightness = -1.0
			}

			if eqBrightness > 1.0 {
				eqBrightness = 1.0
			}

			filters = append(filters, fmt.Sprintf("eq=brightness=%.4f", eqBrightness))
		case "RESIZE":
			if scaleStr, ok := effect.Params["scale"]; ok {
				scale, err := strconv.ParseFloat(scaleStr, 64)
				if err == nil && scale > 0 {
					filters = append(filters, fmt.Sprintf("scale=trunc(iw*%.4f):trunc(ih*%.4f)", scale, scale))
				}
			} else {
				w := effect.Params["width"]
				h := effect.Params["height"]

				if w != "" && h != "" {
					filters = append(filters, fmt.Sprintf("scale=%s:%s", w, h))
				}
			}
		case "CONTRAST":
			factor := 1.0
			if f, err := strconv.ParseFloat(effect.Params["factor"], 64); err == nil {
				factor = f
			}

			filters = append(filters, fmt.Sprintf("eq=contrast=%.4f", factor))
		case "SEPIA":
			intensity := 1.0
			if f, err := strconv.ParseFloat(effect.Params["intensity"], 64); err == nil {
				intensity = f
			}

			// colorchannelmixer takes a 4x4 matrix: R_out, G_out, B_out, A_out
			// Identity: 1 0 0 0 | 0 1 0 0 | 0 0 1 0 | 0 0 0 1
			// Sepia:    0.393 0.769 0.189 0 | 0.349 0.686 0.168 0 | 0.272 0.534 0.131 0 | 0 0 0 1
			// Interpolate each coefficient toward identity as intensity → 0
			ri := 1*(1-intensity) + 0.393*intensity
			gi := 0*(1-intensity) + 0.769*intensity
			bi := 0*(1-intensity) + 0.189*intensity
			rr := 0*(1-intensity) + 0.349*intensity
			gr := 1*(1-intensity) + 0.686*intensity
			br := 0*(1-intensity) + 0.168*intensity
			rb := 0*(1-intensity) + 0.272*intensity
			gb := 0*(1-intensity) + 0.534*intensity
			bb := 1*(1-intensity) + 0.131*intensity
			filters = append(filters, fmt.Sprintf(
				"colorchannelmixer=%.4f:%.4f:%.4f:0:%.4f:%.4f:%.4f:0:%.4f:%.4f:%.4f:0:0:0:0:1",
				ri, gi, bi, rr, gr, br, rb, gb, bb))
		case "VIGNETTE":
			intensity := 1.0
			if f, err := strconv.ParseFloat(effect.Params["intensity"], 64); err == nil {
				intensity = f
			}

			filters = append(filters, fmt.Sprintf("vignette=PI/%.4f", 4.0/intensity))
		}
	}

	if len(filters) == 0 {
		return "", nil
	}

	return strings.Join(filters, ","), nil
}

// ProcessVideoWithFilters applies an ffmpeg filter chain to a video file and writes the result to outputPath.
func ProcessVideoWithFilters(ctx context.Context, inputPath, outputPath, filterChain string) error {
	ffmpegThreads := os.Getenv("FFMPEG_THREADS")
	if ffmpegThreads == "" {
		ffmpegThreads = "2"
	}

	args := []string{"-y", "-threads", ffmpegThreads, "-i", inputPath}

	if filterChain != "" {
		args = append(args, "-vf", filterChain)
	}

	args = append(args, "-c:v", "libx264", "-pix_fmt", "yuv420p", outputPath)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg process: %w: %s", err, stderr.String())
	}

	return nil
}

func getVideoMetadata(ctx context.Context, videoPath string) (int, int, float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate",
		"-of", "csv=p=0", videoPath)

	var out bytes.Buffer

	cmd.Stdout = &out

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, 0, 0, fmt.Errorf("ffprobe: %w: %s", err, stderr.String())
	}

	line := strings.TrimSpace(out.String())

	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return 0, 0, 0, fmt.Errorf("unexpected ffprobe output: %s", line)
	}

	width, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	height, _ := strconv.Atoi(strings.TrimSpace(parts[1]))

	fpsParts := strings.Split(strings.TrimSpace(parts[2]), "/")

	var fps float64

	if len(fpsParts) == 2 {
		num, _ := strconv.ParseFloat(fpsParts[0], 64)

		den, _ := strconv.ParseFloat(fpsParts[1], 64)
		if den > 0 {
			fps = num / den
		}
	}

	return width, height, fps, nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// SplitVideoIntoSegments splits a video file into smaller MP4 segments.
func SplitVideoIntoSegments(ctx context.Context, videoPath, outputDir string, segmentTimeSec int) ([]string, error) {
	if err := ensureDir(outputDir); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	pattern := filepath.Join(outputDir, "part_%03d.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", videoPath, "-c", "copy", "-map", "0", "-segment_time", strconv.Itoa(segmentTimeSec), "-f", "segment", pattern)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg split: %w: %s", err, stderr.String())
	}

	matches, err := filepath.Glob(filepath.Join(outputDir, "part_*.mp4"))
	if err != nil {
		return nil, fmt.Errorf("glob segments: %w", err)
	}

	return matches, nil
}
