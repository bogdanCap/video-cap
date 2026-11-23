package worker

import (
    "fmt"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/detection"
)

type CameraFrameWorkerConfig struct {
    WorkerType camera.CameraType
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

    case camera.USBType:
        
        return NewCameraFrameModifyWorker(
            config.Cam,
			config.Detector,
        ), nil

    case camera.CSIType:
	
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
