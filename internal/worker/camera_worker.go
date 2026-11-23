package worker

import (
	"context"
	//"image"

	"github.com/bogdanCap/video-cap/internal/detection"
	"github.com/bogdanCap/video-cap/internal/camera"
)

type CameraWorker interface {
	Run(cancelCtx context.Context, chanIsFaceDetect <-chan bool) (<-chan camera.FrameProcessDTO, <-chan detection.Face) 
	Wait()
}


