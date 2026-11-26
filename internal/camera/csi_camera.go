package camera

import (
	"bufio"
	//"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"image"
	//"time"
	//"image/jpeg"
	//"bytes"
	//"log"
)

type CSICamera struct {
	cmd    *exec.Cmd
	reader *bufio.Reader
	stdout io.ReadCloser
	scanner *bufio.Scanner

	width     int
	height    int
	frameSize int

	mu sync.Mutex

	fpsFrames int
    //fpsStart  time.Time
}

func NewCSICamera(
	width int,
	height int,
	fps uint32,
) (*CSICamera, error) {
	if width%2 != 0 || height%2 != 0 {
		return nil, fmt.Errorf("width and height must be even for YUV420")
	}

	//TODO mjpeg replace with YUV420 codec

	/* old work
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
	)*/
	cmd := exec.Command(
		"rpicam-vid",
		"-t", "0",
		"-n",
		"--codec", "yuv420",
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

	/*
	scanner := bufio.NewScanner(stdout)
	// Allocate a 2MB max buffer for large JPEG frames
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	scanner.Split(splitJPEG)

	return &CSICamera{
		cmd:    cmd,
		reader: bufio.NewReader(stdout),
		stdout: stdout,
		//fpsStart: time.Now(),
	}, nil*/


	
	//frameSize := width * height * 3 / 2

	yStride := (width + 63) &^ 63
	alignedHeight := (height + 15) &^ 15
	cStride := yStride / 2

	ySize := yStride * alignedHeight
	uvSize := cStride * (alignedHeight / 2)
	frameSize := ySize + (2 * uvSize)

	return &CSICamera{
		cmd:       cmd,
		stdout:    stdout,
		width:     width,
		height:    height,
		frameSize: frameSize,
	}, nil
}

// splitJPEG uses SIMD-accelerated bytes.Index instead of byte-by-byte loops
/*
func splitJPEG(data []byte, atEOF bool) (advance int, token []byte, err error) {
	start := bytes.Index(data, []byte{0xFF, 0xD8})
	if start == -1 {
		if atEOF {
			return len(data), nil, nil
		}
		// Keep last byte in case 0xFF is at the very end of the chunk
		if len(data) > 0 && data[len(data)-1] == 0xFF {
			return len(data) - 1, nil, nil
		}
		return len(data), nil, nil
	}

	end := bytes.Index(data[start+2:], []byte{0xFF, 0xD9})
	if end == -1 {
		if atEOF {
			return len(data), nil, nil
		}
		return start, nil, nil // Need more data
	}

	frameEnd := start + 2 + end + 2
	return frameEnd, data[start:frameEnd], nil
}*/

// Read returns one JPEG frame.
func (c *CSICamera) Read() (FrameProcessDTO, error) {

	/*
	var frame bytes.Buffer

	// Find JPEG Start Of Image (FF D8).
	for {
		b, err := c.reader.ReadByte()
		if err != nil {
			return FrameProcessDTO{}, fmt.Errorf("read JPEG start: %w", err)
		}

		if b != 0xFF {
			continue
		}

		next, err := c.reader.ReadByte()
		if err != nil {
			return FrameProcessDTO{}, fmt.Errorf("read JPEG start marker: %w", err)
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
			return FrameProcessDTO{}, fmt.Errorf("read JPEG frame: %w", err)
		}

		frame.WriteByte(b)

		if prev == 0xFF && b == 0xD9 {
			break
		}

		prev = b
	}
	
	

	// Clone raw bytes safely so FFmpeg worker can consume them concurrently
	jpegBytes := bytes.Clone(frame.Bytes())*/


	/*
	//new
	if !c.scanner.Scan() {
		if err := c.scanner.Err(); err != nil {
			return FrameProcessDTO{}, fmt.Errorf("scan JPEG frame: %w", err)
		}
		return FrameProcessDTO{}, io.EOF
	}

	// Copy exact frame bytes out of scanner's reusable buffer
	raw := c.scanner.Bytes()
	jpegBytes := make([]byte, len(raw))
	copy(jpegBytes, raw)
	////


	// Decode once for Fyne preview
	img, err := jpeg.Decode(bytes.NewReader(jpegBytes))
	if err != nil {
		return FrameProcessDTO{}, fmt.Errorf("decode JPEG frame: %w", err)
	}

	//return frame.Bytes(), nil

	return FrameProcessDTO{
		JPEGBytes:  jpegBytes,
		Image: img,
	}, nil*/

	
	buf := make([]byte, c.frameSize)

	// Read exact frame size in one go
	if _, err := io.ReadFull(c.stdout, buf); err != nil {
		return FrameProcessDTO{}, fmt.Errorf("read YUV420 frame: %w", err)
	}

	ySize := c.width * c.height
	uvSize := ySize / 4

	// Zero-copy wrap of the raw bytes into Go's native image.YCbCr
	img := &image.YCbCr{
		Y:              buf[:ySize],
		Cb:             buf[ySize : ySize+uvSize],
		Cr:             buf[ySize+uvSize:],
		YStride:        c.width,
		CStride:        c.width / 2,
		SubsampleRatio: image.YCbCrSubsampleRatio420,
		Rect:           image.Rect(0, 0, c.width, c.height),
	}

	return FrameProcessDTO{
		JPEGBytes: buf,
		Image:  img,
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