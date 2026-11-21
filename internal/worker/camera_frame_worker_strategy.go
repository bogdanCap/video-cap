package worker

import (
    "fmt"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/detection"
)

type workerType uint8

//TODO csi and usb type move to one Enum in one place
const (
    ModifyType workerType = 1 // can modify frame (face, rotation, etc) for usb
    CleanType workerType = 2 //do not modify frame for csi
)

type CameraFrameWorkerConfig struct {
    WorkerType workerType
    Cam camera.Camera
	Detector detection.Detector
}

type CameraFrameWorkerStrategy struct{}

func NewCameraFrameWorkerStrategy() *CameraFrameWorkerStrategy {
    return &CameraFrameWorkerStrategy{}
}

func (f *CameraFrameWorkerStrategy) Create(
    config CameraFrameWorkerConfig,
) (CameraWorker, error) {

    switch config.WorkerType{

    case ModifyType:
        return NewCameraFrameModifyWorker(
            config.Cam,
			config.Detector,
        ), nil

    case CleanType:
		
        return NewCameraFrameCleanWorker(
            config.Cam,
			config.Detector,
        ), nil

    default:
        return nil, fmt.Errorf(
            "unsupported camera worker type: %s",
            config.WorkerType,
        )
    }
}
