package worker

import (
	"context"
	"log"
	"sync"
	"image/jpeg"
	"bytes"
	"time"
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
			lastFace    *detection.Face
			isFaceDetect bool
		)

		faceDetectionTicker := time.NewTicker(100 * time.Millisecond)
		defer faceDetectionTicker.Stop()

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

			if err != nil {
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

			/*
			quality: 320 - 640 - 720 - 960 - 1024 - 1280
			this need if we need reduce image size - but library already can do it in main.go 
			*/
			//reduceFrame, origWidth, origHeight, targetWidth, targetHeight := camera.ReduceImage(imgFrame, 640)
			//flipFrame, flipError := camera.FlipImage(reduceFrame)

			flipFrame, flipError := camera.FlipImage(imgFrame)

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
			isFaceDetectInterval := false
			select {
			case <-faceDetectionTicker.C:
				isFaceDetectInterval = true
			default:
				// Ticker hasn't fired yet, move on immediately
			}

			if isFaceDetect && isFaceDetectInterval {
				faces, err := w.detector.Detect(flipFrame/*, origWidth, origHeight, targetWidth, targetHeight*/)

				//TODO need to add if after last face detection expire 2 seconds and nothing find - need to reset lastFace
				//this will remove last draw cross from the screen 

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
				} 
				
				// Draw the last detected face.
				if lastFace != nil {
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

			if !isFaceDetect {
				lastFace = nil
			}

			//this help to draw face on every frame using previous object
			if lastFace != nil {
				//show last detected face
				flipFrame, err = w.detector.DrawFaces(
						flipFrame,
						[]detection.Face{*lastFace},
					)

				if err != nil {
					log.Println("draw face:", err)
					continue
				}
			}

			var outBuf bytes.Buffer
			completFrameByteErr := jpeg.Encode(&outBuf, flipFrame, &jpeg.Options{Quality: 90})

			if completFrameByteErr != nil {
				log.Println("camera frame flip convert to byte: ", completFrameByteErr)
				
				return
			}
			completeFrame := outBuf.Bytes()

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

	//TODO for better video quality better to return 2 frame - one with reduce size video and second with original quality
	return frameChan, nil
}

func (w *CameraWorker) Wait() {
	w.wg.Wait()
}
