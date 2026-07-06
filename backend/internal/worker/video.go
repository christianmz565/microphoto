package worker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// FrameInfo holds the path and index of an extracted video frame.
type FrameInfo struct {
	Path  string
	Index int
}

// ExtractFrames extracts all frames from a video file as JPEG images.
func ExtractFrames(ctx context.Context, videoPath, outputDir string) ([]FrameInfo, int, int, float64, error) {
	if err := ensureDir(outputDir); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("create output dir: %w", err)
	}

	width, height, fps, err := getVideoMetadata(ctx, videoPath)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("get metadata: %w", err)
	}

	ffmpegThreads := os.Getenv("FFMPEG_THREADS")
	if ffmpegThreads == "" {
		ffmpegThreads = "2"
	}

	pattern := filepath.Join(outputDir, "frame_%06d.jpg")
	cmd := exec.CommandContext(ctx, "ffmpeg", "-threads", ffmpegThreads, "-i", videoPath, "-q:v", "2", "-fps_mode", "passthrough", pattern)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("ffmpeg extract: %w: %s", err, stderr.String())
	}

	matches, err := filepath.Glob(filepath.Join(outputDir, "frame_*.jpg"))
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("glob frames: %w", err)
	}

	frames := make([]FrameInfo, len(matches))
	for i, m := range matches {
		frames[i] = FrameInfo{Path: m, Index: i}
	}

	return frames, width, height, fps, nil
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

// ReassembleVideo combines processed frames back into a video.
func ReassembleVideo(ctx context.Context, frameDir, outputVideoPath string, fps float64) error {
	pattern := filepath.Join(frameDir, "frame_%06d.jpg")

	ffmpegThreads := os.Getenv("FFMPEG_THREADS")
	if ffmpegThreads == "" {
		ffmpegThreads = "1"
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y",
		"-threads", ffmpegThreads,
		"-framerate", fmt.Sprintf("%.3f", fps),
		"-i", pattern,
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		outputVideoPath)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg reassemble: %w: %s", err, stderr.String())
	}

	return nil
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
