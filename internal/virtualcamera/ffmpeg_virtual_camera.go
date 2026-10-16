package virtualcamera

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type FFmpegVirtualCamera struct {
	devicePath string
	width      int
	height     int
	fps        int

	cmd  *exec.Cmd
	stdin io.WriteCloser

	mu sync.Mutex
}

func NewFFmpegVirtualCamera(
	devicePath string,
	width int,
	height int,
	fps int,
) *FFmpegVirtualCamera {
	return &FFmpegVirtualCamera{
		devicePath: devicePath,
		width:      width,
		height:     height,
		fps:        fps,
	}
}

func (c *FFmpegVirtualCamera) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd != nil {
		return fmt.Errorf("virtual camera is already started")
	}

	if c.devicePath == "" {
		return fmt.Errorf("virtual camera device path is empty")
	}

	if c.width <= 0 || c.height <= 0 || c.fps <= 0 {
		return fmt.Errorf("invalid virtual camera configuration")
	}

	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "warning",

		// Input: concatenated JPEG frames.
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-framerate", fmt.Sprint(c.fps),
		"-i", "pipe:0",

		// Output: virtual V4L2 camera.
		"-vf", fmt.Sprintf(
			"scale=%d:%d,format=yuv420p",
			c.width,
			c.height,
		),
		"-f", "v4l2",
		/*
		"-vf", fmt.Sprintf("scale=%d:%d,format=yuv420p", c.width, c.height),
		"-vcodec", "rawvideo", // 👈 Crucial: Must explicitly state rawvideo for v4l2!
		"-f", "v4l2",*/
		c.devicePath,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create FFmpeg stdin: %w", err)
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		return fmt.Errorf("start FFmpeg: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin

	return nil
}

func (c *FFmpegVirtualCamera) WriteFrame(frame []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil || c.stdin == nil {
		return fmt.Errorf("virtual camera is not started")
	}

	if len(frame) == 0 {
		return fmt.Errorf("empty frame")
	}

	// Write the entire JPEG frame.
	for len(frame) > 0 {
		n, err := c.stdin.Write(frame)
		if err != nil {
			return fmt.Errorf("write frame to FFmpeg: %w", err)
		}

		if n == 0 {
			return fmt.Errorf("write frame to FFmpeg: zero bytes written")
		}

		frame = frame[n:]
	}

	return nil
}

func (c *FFmpegVirtualCamera) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil {
		return nil
	}

	cmd := c.cmd
	stdin := c.stdin

	// Clear the state so subsequent calls see a stopped camera.
	c.cmd = nil
	c.stdin = nil

	if err := stdin.Close(); err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("close FFmpeg stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("stop FFmpeg: %w", err)
	}

	return nil
}
