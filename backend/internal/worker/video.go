package worker

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// ExtractFrames extracts all frames from a video file as PNG images.
// Returns the list of frame paths in MinIO and frame metadata.
type FrameInfo struct {
	Path string
	Index int
}

func ExtractFrames(videoPath string, outputDir string) ([]FrameInfo, int, int, float64, error) {
	if err := ensureDir(outputDir); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("create output dir: %w", err)
	}

	// Get video metadata
	width, height, fps, err := getVideoMetadata(videoPath)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("get metadata: %w", err)
	}

	// Extract frames as PNG
	pattern := filepath.Join(outputDir, "frame_%06d.png")
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vsync", "0", pattern)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("ffmpeg extract: %w: %s", err, stderr.String())
	}

	// Count extracted frames
	matches, err := filepath.Glob(filepath.Join(outputDir, "frame_*.png"))
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("glob frames: %w", err)
	}

	frames := make([]FrameInfo, len(matches))
	for i, m := range matches {
		frames[i] = FrameInfo{Path: m, Index: i}
	}

	return frames, width, height, fps, nil
}

func getVideoMetadata(videoPath string) (int, int, float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error",
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

	// Output format: "width,height,fps_num/fps_den"
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
func ReassembleVideo(frameDir string, outputVideoPath string, fps float64) error {
	pattern := filepath.Join(frameDir, "frame_%06d.png")

	cmd := exec.Command("ffmpeg", "-y",
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
	return exec.Command("mkdir", "-p", dir).Run()
}

func SplitVideoIntoSegments(videoPath string, outputDir string, segmentTimeSec int) ([]string, error) {
	if err := ensureDir(outputDir); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	pattern := filepath.Join(outputDir, "part_%03d.mp4")
	cmd := exec.Command("ffmpeg", "-y", "-i", videoPath, "-c", "copy", "-map", "0", "-segment_time", strconv.Itoa(segmentTimeSec), "-f", "segment", pattern)
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
