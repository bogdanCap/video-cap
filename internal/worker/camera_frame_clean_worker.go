package worker

import (
	"context"
	"log"
	"sync"
	"image/jpeg"
	"bytes"
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
			imgFrame := frameDTO.Image

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

			/*
			imgFrame, err := jpeg.Decode(bytes.NewReader(rawFrame))

			if err != nil {
				log.Println("decode JPEG frame: %w", err)

				return
			}*/

			/*
			quality: 320 - 640 - 720 - 960 - 1024 - 1280
			this need if we need reduce image size - but library already can do it in main.go 
			*/
			//reduceFrame, origWidth, origHeight, targetWidth, targetHeight := camera.ReduceImage(imgFrame, 640)
			//flipFrame, flipError := camera.FlipImage(reduceFrame)

			//flipFrame, flipError := camera.FlipImage(imgFrame)
			//originFrame := flipFrame

			//flipFrame := imgFrame
			
			/*
			if flipError != nil {
				log.Println("camera frame flip: ", err)
				
				return
			}*/
			
			//TODO add select to listen channel that stop button is clicked 
			//and add empty frame for preview
			/*
				width := 640
				height := 480

				frame := make([]byte, width*height*3)
			*/

			
			//isFaceDetectInterval := false
			select {
			case <-faceDetectionTicker.C:
				//isFaceDetectInterval = true
			default:
				// Ticker hasn't fired yet, move on immediately
			}
			
			/*
			if isFaceDetect && isFaceDetectInterval {
				faces, err := w.detector.Detect(flipFrame)

				//TODO need to add if after last face detection expire 2 seconds and nothing find - need to reset lastFace
				//this will remove last draw cross from the screen 

				if err != nil {
					log.Println("face detection:", err)

				} else if len(faces) > 0 {
					face := &faces[0]
					lastFace = face
				} 
				
				// Draw the last detected face.
				if lastFace != nil {
					//logic from face_service placed here
					//TODO in this place i think need to add Struct with property 


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
			*/

			//TODO this need to to outside and return flipFrame
			
			
			var outBuf bytes.Buffer
			completFrameByteErr := jpeg.Encode(
				&outBuf, 
				imgFrame, 
				&jpeg.Options{Quality: 75},
			)

			if completFrameByteErr != nil {
				log.Println("camera frame flip convert to byte: ", completFrameByteErr)
				
				return
			}
			completeFrame := outBuf.Bytes()

			//TODO instead frameChan need to use Struct with property originFrame (foriginFrame) and face (lastFace)
			//TODO this need for face_service lgic

			// Send frame to the channel.
			// usb logic
			/*
			select {
			case frameChan <- completeFrame:
			//case faceImageChan <- *lastFace:	

			case <-cancelCtx.Done():
				log.Println("camera worker stopped")

				return
			//TODO old default change with new one
			default:
				// The massive 30-frame buffer is 100% full! Drop immediately to prevent infinite memory growth
    			log.Println("[CameraWorker] Buffer completely full! Dropping frame.")
			}*/
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
				frameCaptureChan <- camera.FrameProcessDTO{JPEGBytes: completeFrame, Image: imgFrame}//completeFrame
			}
		}
	}()

	//TODO for better video quality better to return 2 frame - one with reduce size video and second with original quality
	return frameCaptureChan, nil
}

func (w *CameraFrameCleanWorker) Wait() {
	w.wg.Wait()
}
