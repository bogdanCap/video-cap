package camera

import (
    "fmt"
)

type CameraConfig struct {
    Type        CameraType
    DevicePath  string
    Width       int
    Height      int
    FPS         uint32

    //this is only for csi_dual
    PreviewWidth,
	PreviewHeight int
	ChunkSeconds int
}

type CameraStrategy struct{}

func NewCameraStrategy() *CameraStrategy {
    return &CameraStrategy{}
}

func (f *CameraStrategy) Create(
    config CameraConfig,
) (Camera, error) {

    switch config.Type {

    case CSIType:
        return NewCSICamera(
            config.Width,
            config.Height,
            config.FPS,
        )
    case CSIDualType:
        return NewCSIDualPipeline(
            CSIDualConfig{
                outputDir: "recordings",
                recWidth: config.Width,
                recHeight: config.Height,
                fps: config.FPS,
                previewWidth: config.PreviewWidth,
                previewHeight: config.PreviewHeight,
                chunkSeconds: config.ChunkSeconds,
            },
        )
    case USBType:
		if config.DevicePath == "" {
            return nil, fmt.Errorf("USB camera device path is required")
        }

        return NewV4LCamera(
            config.DevicePath,
            config.Width,
            config.Height,
            config.FPS,
        )

    default:
        return nil, fmt.Errorf(
            "unsupported camera type: %s",
            config.Type,
        )
    }
}
