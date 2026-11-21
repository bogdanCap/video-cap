package camera

import (
	"image"
)

// Camera describes a camera capable of streaming frames.
type Camera interface {
	Read() (FrameRead, error)
	Stop() error
	Close() error
}

type FrameRead struct {
	JPEGBytes []byte      // Raw high-res JPEG bytes -> pipe directly to FFmpeg
	Image image.Image // Decoded image -> pass directly to FynePreview
}
