package camera

import (
	"image"
)

// Camera describes a camera capable of streaming frames.
type Camera interface {
	Read() (FrameProcessDTO, error)
	Stop() error
	Close() error
}

//TODO Need one DTO FrameRead rename to CameraFrameDTO
type FrameProcessDTO struct {
	JPEGBytes []byte      // Raw high-res JPEG bytes -> pipe directly to FFmpeg
	Image image.Image // Decoded image -> pass directly to FynePreview
}
