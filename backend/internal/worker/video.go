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
