package camera

import (
    "fmt"
)

/*
type CameraType string

const (
    CameraTypeCSI CameraType = "csi"
    CameraTypeUSB CameraType = "usb"
)*/

type CameraConfig struct {
    Type        CameraType
    DevicePath  string
    Width       int
    Height      int
    FPS         uint32
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
