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
)

type CSICamera struct {
	cmd    *exec.Cmd
	reader *bufio.Reader
	stdout io.ReadCloser

	mu sync.Mutex
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
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--framerate", strconv.FormatUint(uint64(fps), 10),
		"-o", "-",
	)

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
	}, nil
}

// Read returns one JPEG frame.
func (c *CSICamera) Read() ([]byte, error) {

	var frame bytes.Buffer

	// Find JPEG Start Of Image (FF D8).
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read JPEG start: %w", err)
		}

		if b != 0xFF {
			continue
		}

		next, err := c.reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read JPEG start marker: %w", err)
		}

		if next == 0xD8 {
			frame.Write([]byte{0xFF, 0xD8})
			break
		}
	}

	// Find JPEG End Of Image (FF D9).
	prev := byte(0)

	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("read JPEG frame: %w", err)
		}

		frame.WriteByte(b)

		if prev == 0xFF && b == 0xD9 {
			break
		}

		prev = b
	}

	return frame.Bytes(), nil
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