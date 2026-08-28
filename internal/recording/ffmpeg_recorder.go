package recording

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type FFmpegRecorder struct {
	outputDir string

	width  int
	height int
	fps    int

	chunkDuration time.Duration

	input io.WriteCloser
	cmd   *exec.Cmd

	chunkStart time.Time
}

func NewFFmpegRecorder(
	outputDir string,
	width int,
	height int,
	fps int,
	chunkDuration time.Duration,
) *FFmpegRecorder {

	return &FFmpegRecorder{
		outputDir:     outputDir,
		width:         width,
		height:        height,
		fps:           fps,
		chunkDuration: chunkDuration,
	}
}

func (r *FFmpegRecorder) Start() error {

	if r.input != nil {
		return fmt.Errorf("recorder is already running")
	}

	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf(
			"create recording directory: %w",
			err,
		)
	}

	return r.startChunk()
}

func (r *FFmpegRecorder) startChunk() error {

	timestamp := time.Now().Format(
		"20060102_150405",
	)

	filename := filepath.Join(
		r.outputDir,
		fmt.Sprintf(
			"video_%s.mp4",
			timestamp,
		),
	)

	cmd := exec.Command(
		"ffmpeg",

		"-y",

		// Input is MJPEG.
		"-f", "mjpeg",

		// Input FPS.
		"-framerate",
		fmt.Sprintf("%d", r.fps),

		// Read MJPEG frames from stdin.
		"-i", "pipe:0",

		// Output resolution.
		"-vf",
		fmt.Sprintf(
			"scale=%d:%d",
			r.width,
			r.height,
		),

		// H.264.
		"-c:v", "libx264",

		// Compatibility.
		"-pix_fmt", "yuv420p",

		// MP4 optimization.
		"-movflags", "+faststart",

		filename,
	)

	input, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf(
			"create ffmpeg stdin: %w",
			err,
		)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		_ = input.Close()

		return fmt.Errorf(
			"start ffmpeg: %w",
			err,
		)
	}

	r.input = input
	r.cmd = cmd
	r.chunkStart = time.Now()

	fmt.Printf(
		"Recording started: %s\n",
		filename,
	)

	return nil
}

func (r *FFmpegRecorder) WriteFrame(
	frame []byte,
) error {

	if r.input == nil {
		return fmt.Errorf(
			"recorder is not running",
		)
	}

	_, err := r.input.Write(frame)
	if err != nil {
		return fmt.Errorf(
			"write frame to ffmpeg: %w",
			err,
		)
	}

	if time.Since(r.chunkStart) >= r.chunkDuration {
		return r.rotateChunk()
	}

	return nil
}

func (r *FFmpegRecorder) rotateChunk() error {

	if err := r.closeChunk(); err != nil {
		return err
	}

	return r.startChunk()
}

func (r *FFmpegRecorder) closeChunk() error {

	if r.input == nil {
		return nil
	}

	// EOF for FFmpeg.
	if err := r.input.Close(); err != nil {
		r.input = nil

		return fmt.Errorf(
			"close ffmpeg input: %w",
			err,
		)
	}

	err := r.cmd.Wait()

	r.input = nil
	r.cmd = nil

	if err != nil {
		return fmt.Errorf(
			"ffmpeg: %w",
			err,
		)
	}

	fmt.Println("Recording chunk saved")

	return nil
}

func (r *FFmpegRecorder) Stop() error {
	return r.closeChunk()
}
