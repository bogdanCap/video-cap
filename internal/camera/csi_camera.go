package camera

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
	"image/jpeg"
	//"log"
)

type CSICamera struct {
	cmd    *exec.Cmd
	reader *bufio.Reader
	stdout io.ReadCloser

	mu sync.Mutex

	fpsFrames int
    fpsStart  time.Time
}

func NewCSICamera(
	width int,
	height int,
	fps uint32,
) (*CSICamera, error) {

	cmd := exec.Command(
		"rpicam-vid",
		"-t", "0",
		"-n",
		"--codec", "mjpeg",
		"--quality", "75", // Cuts JPEG []byte size in half -> faster jpeg.Decode()
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--framerate", strconv.FormatUint(uint64(fps), 10),
		"-o", "-",
	)

	/* check this
	rpicam := exec.Command(
		"rpicam-vid",
		"-t", "0",          // Record until stopped
		"-n",               // Disable built-in fullscreen preview window
		"--width", "1920",
		"--height", "1080",
		"--framerate", "30",
		"--inline",         // Required for clean stream splitting
		"-o", "-",          // Output to stdout
	)

	"rpicam-vid",
	"-t", "0",
	"-n",
	"--codec", "mjpeg",
	"--quality", "75", // Cuts JPEG []byte size in half -> faster jpeg.Decode()
	"--width", strconv.Itoa(width),
	"--height", strconv.Itoa(height),
	"--framerate", strconv.FormatUint(uint64(fps), 10),
	"-o", "-",
	*/

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		stdout.Close()
		return nil, fmt.Errorf("start CSI camera: %w", err)
	}

	return &CSICamera{
		cmd:    cmd,
		reader: bufio.NewReader(stdout),
		stdout: stdout,
		fpsStart: time.Now(),
	}, nil
}

// Read returns one JPEG frame.
func (c *CSICamera) Read() (FrameRead, error) {

	var frame bytes.Buffer

	// Find JPEG Start Of Image (FF D8).
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return FrameRead{}, fmt.Errorf("read JPEG start: %w", err)
		}

		if b != 0xFF {
			continue
		}

		next, err := c.reader.ReadByte()
		if err != nil {
			return FrameRead{}, fmt.Errorf("read JPEG start marker: %w", err)
		}

		if next == 0xD8 {
			frame.Write([]byte{0xFF, 0xD8})
			break
		}
	}

	// Read until JPEG End Of Image (FF D9)
	prev := byte(0)

	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return FrameRead{}, fmt.Errorf("read JPEG frame: %w", err)
		}

		frame.WriteByte(b)

		if prev == 0xFF && b == 0xD9 {
			break
		}

		prev = b
	}

	// FPS measurement.
	/*
    c.fpsFrames++

    elapsed := time.Since(c.fpsStart)

    if elapsed >= time.Second {
        fps := float64(c.fpsFrames) / elapsed.Seconds()

        log.Printf("CSI Camera FPS: %.2f", fps)

        c.fpsFrames = 0
        c.fpsStart = time.Now()
    }*/

	// Clone raw bytes safely so FFmpeg worker can consume them concurrently
	jpegBytes := bytes.Clone(frame.Bytes())

	// Decode once for Fyne preview
	img, err := jpeg.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		return FrameRead{}, fmt.Errorf("decode JPEG frame: %w", err)
	}


	//return frame.Bytes(), nil

	return FrameRead{
		JPEGBytes:  jpegBytes,
		Image: img,
	}, nil
}

// Stop terminates camera capture.
func (c *CSICamera) Stop() error {

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	if err := c.cmd.Process.Kill(); err != nil &&
		!errors.Is(err, os.ErrProcessDone) {
		return fmt.Errorf("stop CSI camera: %w", err)
	}

	return nil
}

// Close releases camera resources.
func (c *CSICamera) Close() error {

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	// Stop process if it is still running.
	_ = c.cmd.Process.Kill()

	// Close stdout pipe.
	_ = c.stdout.Close()

	// Wait for process termination.
	err := c.cmd.Wait()

	c.cmd.Process = nil

	if err != nil {
		return fmt.Errorf("close CSI camera: %w", err)
	}

	return nil
}