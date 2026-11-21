package worker

import (
	"context"
	"image"

	"github.com/bogdanCap/video-cap/internal/detection"
)

type CameraWorker interface {
	Run(cancelCtx context.Context, chanIsFaceDetect <-chan bool) (<-chan FrameCaptureChan, <-chan detection.Face) 
	Wait()
}

type FrameCaptureChan struct {
	JPEGBytes []byte      // Raw high-res JPEG bytes -> pipe directly to FFmpeg
	Image image.Image // Decoded image -> pass directly to FynePreview
}


