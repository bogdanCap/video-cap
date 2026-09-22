package worker

import (
	"context"
	"log"
	"sync"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/detection"
)

type CameraWorker struct {
	cam camera.Camera
	detector detection.Detector
	wg sync.WaitGroup
}

func NewCameraWorker(
	cam camera.Camera,
	detector detection.Detector,
) *CameraWorker {
	return &CameraWorker{
		cam:      cam,
		detector: detector,
	}
}

func (w *CameraWorker) Run(ctx context.Context, chanIsFaceDetect <-chan bool) (<-chan []byte, <-chan detection.Face ) {
	frameChan := make(chan []byte, 30)
	faceImageChan := make(chan detection.Face, 30)

	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		defer close(frameChan)

		var (
			frameNumber int
			lastFace    *detection.Face
			isFaceDetect bool
		)

		for {
			
			select {
			case <-ctx.Done():
				log.Println("camera worker stopped")
				return
			case isFaceDetect = <-chanIsFaceDetect:
				log.Println("event in handle face detection = ", isFaceDetect)

			default:
			}

			frame, err := w.cam.Read()

			if err != nil {
				//log.Println("camera read:", err)
				// Camera can return an error because cam.Stop()
				// was called during shutdown.
				if ctx.Err() != nil {
					log.Println("camera worker stopped")
				} else {
					log.Println("camera read:", err)
				}

				return
			}

			if isFaceDetect {
				frameNumber++

				// Detect face every 5th frame.
				faces, err := w.detector.Detect(frame)


				if err != nil {
					log.Println("face detection:", err)

				} else if len(faces) > 0 {
					face := faces[0]
					lastFace = &face

				} //else {
					//lastFace = nil
				//}

				
				/*
				if frameNumber%5 == 0 {
					faces, err := w.detector.Detect(frame)

					if err != nil {
						log.Println("face detection:", err)

					} else if len(faces) > 0 {
						face := faces[0]
						lastFace = &face

					} else {
						lastFace = nil
					}
				}*/

				// Draw the last detected face.
				if lastFace != nil {
					frame, err = w.detector.DrawFaces(
						frame,
						[]detection.Face{*lastFace},
					)

					if err != nil {
						log.Println("draw face:", err)
						continue
					}
				}
			}

			// Send frame to the channel.
			select {
			case frameChan <- frame:
			case faceImageChan <- *lastFace:	

			case <-ctx.Done():
				log.Println("camera worker stopped")

				return
			}
		}
	}()

	return frameChan, faceImageChan
}

func (w *CameraWorker) Wait() {
	w.wg.Wait()
}
