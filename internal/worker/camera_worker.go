package worker

import (
	"context"
	"log"
	"sync"
	"image/jpeg"
	"bytes"
	//"fmt"

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

func (w *CameraWorker) Run(ctx context.Context, chanIsFaceDetect <-chan bool) (<-chan []byte, <-chan detection.Face) {
	frameChan := make(chan []byte, 30)
	//faceImageChan := make(chan detection.Face, 30)

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

			rawFrame, err := w.cam.Read()

			frameNumber++

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

			imgFrame, err := jpeg.Decode(bytes.NewReader(rawFrame))

			if err != nil {
				log.Println("decode JPEG frame: %w", err)

				return
			}

			reduceFrame, origWidth, origHeight, targetWidth, targetHeight := camera.ReduceImage(imgFrame, 640)

			flipFrame, flipError := camera.FlipImage(reduceFrame)
			//flipFrame := reduceFrame

			
			if flipError != nil {
				log.Println("camera frame flip: ", err)
				
				return
			}
			
			//TODO add select to listen channel that stop button is clicked 
			//and add empty frame for preview
			/*
				width := 640
				height := 480

				frame := make([]byte, width*height*3)
			*/
			if isFaceDetect && frameNumber%2 == 0 {

				// Detect face every 5th frame.
				faces, err := w.detector.Detect(flipFrame, origWidth, origHeight, targetWidth, targetHeight)

				//var face *detection.Face

				//if i need to draw face in origin image
				/*
				faces = detection.ScaleFaces(
					faces,
					targetWidth, 
					targetHeight,
					origWidth, 
					origHeight,
					true, // Mirror face coordinates.
				)
				*/

				if err != nil {
					log.Println("face detection:", err)

				} else if len(faces) > 0 {
					face := &faces[0]
					lastFace = face

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
				//if face != nil {
					flipFrame, err = w.detector.DrawFaces(
						flipFrame,
						[]detection.Face{*lastFace},
					)

					if err != nil {
						log.Println("draw face:", err)
						continue
					}
				}
			}

			if frameNumber == 10 {
				frameNumber = 0
			}

			var outBuf bytes.Buffer
			completFrameByteErr := jpeg.Encode(&outBuf, flipFrame, &jpeg.Options{Quality: 90})

			if completFrameByteErr != nil {
				log.Println("camera frame flip convert to byte: ", completFrameByteErr)
				
				return
			}

			completeFrame := outBuf.Bytes()

			//return outBuf.Bytes(), nil

			// Send frame to the channel.
			select {
			case frameChan <- completeFrame:
			//case faceImageChan <- *lastFace:	

			case <-ctx.Done():
				log.Println("camera worker stopped")

				return
			}
		}
	}()

	return frameChan, nil//faceImageChan
}

func (w *CameraWorker) Wait() {
	w.wg.Wait()
}
