package recording

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bogdanCap/video-cap/internal/camera"
)

type FFmpegConfig struct {
	OutputDir string
	Width int
	Height int
	Fps int
	ChunkDuration time.Duration
	CameraType camera.CameraType
}

type FFmpegRecorder struct {
	config FFmpegConfig

	input io.WriteCloser
	cmd   *exec.Cmd
	chunkStart time.Time
}

func NewFFmpegRecorder(
	config FFmpegConfig,
) *FFmpegRecorder {

	return &FFmpegRecorder{
		config: config,
	}
}

func (r *FFmpegRecorder) Start() error {

	if r.input != nil {
		return fmt.Errorf("recorder is already running")
	}

	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		return fmt.Errorf(
			"create recording directory: %w",
			err,
		)
	}

	return r.startChunk()
}

func (r *FFmpegRecorder) IsRunning() bool {
	return r.input != nil
}

func (r *FFmpegRecorder) startChunk() error {
	timestamp := time.Now().Format(
		"20060102_150405",
	)

	filename := filepath.Join(
		r.config.OutputDir,
		fmt.Sprintf(
			"video_%s.mp4",
			timestamp,
		),
	)
	
	//this is for face detection mode
	//recorrds video in 640x480 as set-up in V4LCamera
	//check video details - ffprobe video_20260923_011226.mp4
	//version for USB vam TODO this for USB and CSI need to configure from outside and need to add configDTO for constuctor
	var args []string
	
	switch r.config.CameraType {
	case camera.CSIType:
		args = []string{
			"-y",
			"-f", "rawvideo",
			"-pix_fmt", "yuv420p",
			"-s", fmt.Sprintf("%dx%d", r.config.Width, r.config.Height),
			"-r", fmt.Sprintf("%d", r.config.Fps),
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-preset", "ultrafast", // Best speed/CPU balance on Pi 5 OR veryfast (slower)
			"-pix_fmt", "yuv420p",
			filename,
		}
	case camera.USBType:
		//this is for type camera.USBType
		args = []string{
			"-y",
			// Input is a sequence of MJPEG/JPEG frames.
			//"-probesize", "32K",
			//"-analyzeduration", "0",

			
			"-f", "mjpeg",

			//"-f", "image2pipe",
			// 2. Explicitly tell it that the images inside the pipe are mjpegs
			//"-vcodec", "mjpeg",

			"-framerate", fmt.Sprintf("%d", r.config.Fps),
			"-i", "pipe:0",
			// Video encoding.
			"-c:v", "libx264",
			"-preset", "veryfast", // veryfast -> ultrafast (50% less cpu)
			"-pix_fmt", "yuv420p",
			filename,
		}
	default:
		return fmt.Errorf("ffmpeg recorder only required types: %s, %s", camera.CSIType.String(), camera.USBType.String())
	}
	
	
	/*
	ffmpeg := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "h264",
		"-i", "pipe:0",
		"-c:v", "copy", outputFile, // Output 1: Save file
		"-vf", "scale=640:360,fps=15", // Output 2: Lightweight preview
		"-f", "mjpeg",
		"-q:v", "5",
		"pipe:1",
	)
	*/


	cmd := exec.Command(
		"ffmpeg",
		args...
	)

	input, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf(
			"create ffmpeg stdin: %w",
			err,
		)
	}

	//TODO check without this 2 line
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
		return fmt.Errorf("write frame to ffmpeg: %w", err)
	}

	if time.Since(r.chunkStart) >= r.config.ChunkDuration {
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
