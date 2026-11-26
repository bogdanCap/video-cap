package worker

import (
	"context"
	"log"
	"sync"
	//"image/jpeg"
	//"bytes"
	"time"
	//"image"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/detection"
)

type CameraFrameCleanWorker struct {
	cam camera.Camera
	detector detection.Detector
	wg sync.WaitGroup
}

func NewCameraFrameCleanWorker(
	cam camera.Camera,
	detector detection.Detector,
) *CameraFrameCleanWorker {
	return &CameraFrameCleanWorker{
		cam:      cam,
		detector: detector,
	}
}

func (w *CameraFrameCleanWorker) Run(cancelCtx context.Context, chanIsFaceDetect <-chan bool) (<-chan camera.FrameProcessDTO/*[]byte*/, <-chan detection.Face) {
	//TODO test speed replace 30 length with 2
	//usb logic 30 buffer
	//frameChan := make(chan []byte, 30)
	//csi logic 1 buffer
	//frameChan := make(chan []byte, 1)FrameCaptureChan
	frameCaptureChan := make(chan camera.FrameProcessDTO, 1)
	//TODO for csi camera buffer change from 30 to 1
	//frameChan := make(chan []byte, 2) - this need if we need better live preview image 
	//faceImageChan := make(chan detection.Face, 30)

	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		defer close(frameCaptureChan)

		var (
			//lastFace    *detection.Face
			isFaceDetect bool
		)

		faceDetectionTicker := time.NewTicker(faceDetectionTickerInterval)
		defer faceDetectionTicker.Stop()

		for {

			select {
			case <-cancelCtx.Done():
				log.Println("camera worker stopped")
				return
			case isFaceDetect = <-chanIsFaceDetect:
				log.Println("event in handle face detection = ", isFaceDetect)

			default:
			}

			//rawFrame, err := w.cam.Read()
			frameDTO, err := w.cam.Read()
			//imgFrame := frameDTO.Image

			if err != nil {
				// Camera can return an error because cam.Stop()
				// was called during shutdown.
				if cancelCtx.Err() != nil {
					log.Println("camera worker stopped")
				} else {
					log.Println("camera read:", err)
				}

				return
			}

			select {
			case <-cancelCtx.Done():
				log.Println("camera worker stopped")
				return
			default:
				// Drain any unread stale frame so the newest frame always gets in
				select {
				case <-frameCaptureChan:
				default:
				}
				frameCaptureChan <- frameDTO
			}
		}
	}()

	//TODO for better video quality better to return 2 frame - one with reduce size video and second with original quality
	return frameCaptureChan, nil
}

func (w *CameraFrameCleanWorker) Wait() {
	w.wg.Wait()
}
