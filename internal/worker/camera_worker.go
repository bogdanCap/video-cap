package worker

import (
	"context"
	"log"

	"video/internal/camera"
	"video/internal/detection"
)

type CameraWorker struct {
	cam      camera.Camera
	detector detection.Detector
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

func (w *CameraWorker) Run(ctx context.Context) <-chan []byte {
	frameChan := make(chan []byte, 30)

	go func() {
		defer close(frameChan)

		var (
			frameNumber int
			lastFace    *detection.Face
		)

		for {
			select {
			case <-ctx.Done():
				log.Println("camera worker stopped")
				return

			default:
			}

			frame, err := w.cam.Read()
			if err != nil {
				log.Println("camera read:", err)
				return
			}

			frameNumber++

			// Detect face every 5th frame.
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
			}

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

			// Send frame to the channel.
			select {
			case frameChan <- frame:

			case <-ctx.Done():
				return
			}
		}
	}()

	return frameChan
}
