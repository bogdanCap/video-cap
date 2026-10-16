package worker

import (
	"context"
	"log"
	"sync"

	"github.com/bogdanCap/video-cap/internal/virtualcamera"
)

type VirtualCameraWorker struct {
	camera virtualcamera.VirtualCamera

	mu      sync.Mutex
	enabled bool
}

func NewVirtualCameraWorker(
	camera virtualcamera.VirtualCamera,
) *VirtualCameraWorker {
	return &VirtualCameraWorker{
		camera: camera,
	}
}

// Run should be started once for each application capture session.
// It does not start FFmpeg itself.
func (w *VirtualCameraWorker) Run(
	ctx context.Context,
	frameChan <-chan []byte,
) {
	for {
		select {
		case <-ctx.Done():
			log.Println("virtual camera worker stopped")
			return

		case frame, ok := <-frameChan:
			if !ok {
				return
			}

			w.mu.Lock()

			if w.enabled {
				if err := w.camera.WriteFrame(frame); err != nil {
					log.Printf("virtual camera write: %v", err)
				}
			}

			w.mu.Unlock()
		}
	}
}

// StartOutput starts FFmpeg when the virtual camera button is clicked.
func (w *VirtualCameraWorker) StartOutput() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.enabled {
		return nil
	}

	if err := w.camera.Start(); err != nil {
		return err
	}

	w.enabled = true

	return nil
}

// StopOutput stops FFmpeg without stopping the main camera pipeline.
func (w *VirtualCameraWorker) StopOutput() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.enabled {
		return nil
	}

	// Prevent additional writes before stopping FFmpeg.
	w.enabled = false

	return w.camera.Stop()
}

func (w *VirtualCameraWorker) Enabled() bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.enabled
}
