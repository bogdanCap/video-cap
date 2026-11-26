package camera

import (
	"fmt"
	"image"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

type CSIDualPipeline struct {
	rpicamCmd *exec.Cmd
	ffmpegCmd *exec.Cmd
	stdout    io.ReadCloser

	previewW  int
	previewH  int
	frameSize int

	mu sync.Mutex
}

func NewCSIDualPipeline(
	outputDir string,
	recWidth, recHeight int,
	fps uint32,
	previewW, previewH int,
	chunkSeconds int,
) (*CSIDualPipeline, error) {
	if recWidth%64 != 0 || recHeight%16 != 0 {
		return nil, fmt.Errorf("recWidth must be divisible by 64 and recHeight by 16 for zero-padding YUV420")
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	// 1. rpicam-vid outputs raw YUV420 to stdout
	rpicamCmd := exec.Command(
		"rpicam-vid",
		"-t", "0",
		"-n",
		"--flush",
		"--codec", "yuv420",
		"--width", strconv.Itoa(recWidth),
		"--height", strconv.Itoa(recHeight),
		"--framerate", strconv.FormatUint(uint64(fps), 10),
		"-o", "-",
	)

	rpicamOut, err := rpicamCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("rpicam stdout pipe: %w", err)
	}
	rpicamCmd.Stderr = os.Stderr

	// 2. FFmpeg reads rpicam-vid's stdout directly and splits into:
	//    Output 1: 1-minute MP4 segments on disk
	//    Output 2: Lightweight RGBA frames on pipe:1 for FynePreview
	segmentPattern := filepath.Join(outputDir, "video_%Y%m%d_%H%M%S.mp4")

	/*
	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "rawvideo",
		"-pix_fmt", "yuv420p",
		"-s", fmt.Sprintf("%dx%d", recWidth, recHeight),
		"-r", strconv.FormatUint(uint64(fps), 10),
		"-i", "pipe:0",

		// Output 1: 1-minute MP4 chunks
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-pix_fmt", "yuv420p",
		"-f", "segment",
		"-segment_time", strconv.Itoa(chunkSeconds), // 60 for 1 minute
		"-strftime", "1",
		"-reset_timestamps", "1",
		segmentPattern,

		// Output 2: Tiny RGBA stream for Go FynePreview (and optional horizontal flip: "hflip,scale=...")
		"-vf", fmt.Sprintf("scale=%d:%d,fps=20", previewW, previewH),
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"pipe:1",
	)*/

	ffmpegCmd := exec.Command(
		"ffmpeg",
		"-y",

		// --- INPUT: Raw YUV420 stream from rpicam-vid (pipe:0 / stdin) ---
		"-f", "rawvideo",
		"-pix_fmt", "yuv420p",
		"-s", fmt.Sprintf("%dx%d", recWidth, recHeight),
		"-r", strconv.FormatUint(uint64(fps), 10),
		"-i", "pipe:0",

		// --- OUTPUT 1: 1-minute crash-safe MP4 segments on disk ---
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-g", strconv.FormatUint(uint64(fps), 10), // Keyframe every 1 second
		"-pix_fmt", "yuv420p",
		"-f", "segment",
		"-segment_time", strconv.Itoa(chunkSeconds), // 60 for 1-minute chunks
		"-segment_format_options", "movflags=+frag_keyframe+empty_moov+default_base_moof",
		"-strftime", "1",
		"-reset_timestamps", "1",
		segmentPattern,

		// --- OUTPUT 2: Lightweight RGBA preview stream to Go (pipe:1 / stdout) ---
		"-vf", fmt.Sprintf("scale=%d:%d,fps=20", previewW, previewH),
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"pipe:1",
	)

	// Wire rpicam-vid stdout -> FFmpeg stdin directly in OS kernel
	ffmpegCmd.Stdin = rpicamOut
	ffmpegCmd.Stderr = os.Stderr

	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		_ = rpicamOut.Close()
		return nil, fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}

	if err := rpicamCmd.Start(); err != nil {
		_ = rpicamOut.Close()
		_ = ffmpegOut.Close()
		return nil, fmt.Errorf("start rpicam-vid: %w", err)
	}

	if err := ffmpegCmd.Start(); err != nil {
		_ = rpicamCmd.Process.Kill()
		_ = rpicamOut.Close()
		_ = ffmpegOut.Close()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	// Close parent's copy of rpicamOut now that ffmpegCmd has inherited it
	_ = rpicamOut.Close()

	return &CSIDualPipeline{
		rpicamCmd: rpicamCmd,
		ffmpegCmd: ffmpegCmd,
		stdout:    ffmpegOut,
		previewW:  previewW,
		previewH:  previewH,
		frameSize: previewW * previewH * 4, // 4 bytes per pixel (RGBA)
	}, nil
}

// Read returns one ready-to-render RGBA preview frame.
func (c *CSIDualPipeline) Read() (FrameProcessDTO, error) {
	buf := make([]byte, c.frameSize)

	if _, err := io.ReadFull(c.stdout, buf); err != nil {
		return FrameProcessDTO{}, fmt.Errorf("read preview RGBA frame: %w", err)
	}

	// Zero-copy wrap directly into *image.RGBA!
	img := &image.RGBA{
		Pix:    buf,
		Stride: c.previewW * 4,
		Rect:   image.Rect(0, 0, c.previewW, c.previewH),
	}

	return FrameProcessDTO{
		Image: img,
	}, nil
}

func (c *CSIDualPipeline) Stop() error {
	/*
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.rpicamCmd != nil && c.rpicamCmd.Process != nil {
		_ = c.rpicamCmd.Process.Kill()
	}
	if c.ffmpegCmd != nil && c.ffmpegCmd.Process != nil {
		_ = c.ffmpegCmd.Process.Kill()
	}
	return nil*/
	

	/*
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Stop rpicam-vid cleanly so it flushes remaining frames and closes the pipe (EOF)
	if c.rpicamCmd != nil && c.rpicamCmd.Process != nil {
		_ = c.rpicamCmd.Process.Signal(os.Interrupt)
		_ = c.rpicamCmd.Wait()
		c.rpicamCmd = nil
	}

	// 2. Wait for FFmpeg to receive EOF, flush the remaining frames, and finalize the MP4
	if c.ffmpegCmd != nil {
		_ = c.ffmpegCmd.Wait()
		c.ffmpegCmd = nil
	}

	if c.stdout != nil {
		_ = c.stdout.Close()
		c.stdout = nil
	}

	return nil
	*/

	c.mu.Lock()
	rpicam := c.rpicamCmd
	ffmpeg := c.ffmpegCmd
	stdout := c.stdout
	c.rpicamCmd = nil
	c.ffmpegCmd = nil
	c.stdout = nil
	c.mu.Unlock()

	// 1. Drain any remaining preview bytes from ffmpeg's stdout in the background
	// so FFmpeg never blocks on a full stdout pipe while shutting down.
	if stdout != nil {
		go func() {
			_, _ = io.Copy(io.Discard, stdout)
			_ = stdout.Close()
		}()
	}

	// 2. Stop rpicam-vid immediately. Because rpicam-vid only outputs raw YUV420,
	// killing it immediately closes its stdout pipe and sends a clean EOF to FFmpeg's stdin!
	if rpicam != nil && rpicam.Process != nil {
		_ = rpicam.Process.Kill()
		_ = rpicam.Wait()
	}

	// 3. Now that FFmpeg received EOF on stdin and its stdout is draining,
	// FFmpeg flushes the last frames, writes the MP4 trailer, and exits in ~100ms.
	if ffmpeg != nil {
		_ = ffmpeg.Wait()
	}

	return nil
}

func (c *CSIDualPipeline) Close() error {
	_ = c.Stop()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.stdout != nil {
		_ = c.stdout.Close()
	}
	if c.rpicamCmd != nil {
		_ = c.rpicamCmd.Wait()
	}
	if c.ffmpegCmd != nil {
		_ = c.ffmpegCmd.Wait()
	}
	return nil
}
