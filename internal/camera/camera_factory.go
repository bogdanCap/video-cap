package camera

import (
    "fmt"
)

type CameraType string

const (
    CameraTypeCSI CameraType = "csi"
    CameraTypeUSB CameraType = "usb"
)

type CameraConfig struct {
    Type        CameraType
    DevicePath  string
    Width       int
    Height      int
    FPS         uint32
}

type CameraFactory struct{}

func NewCameraFactory() *CameraFactory {
    return &CameraFactory{}
}

func (f *CameraFactory) Create(
    config CameraConfig,
) (Camera, error) {

    switch config.Type {

    case CameraTypeCSI:
        return NewCSICamera(
            config.Width,
            config.Height,
            config.FPS,
        )

    case CameraTypeUSB:
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
