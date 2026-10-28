package service

import (
	"log"
	"context"
	"time"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/motiontracker"
	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/worker"
	"github.com/bogdanCap/video-cap/internal/ui/component"
)

type VideoControlService struct {
    // UI
	//button      *widget.Button
	startButton      *component.StartButton
	stopButton       *component.StopButton
	faceDetectButton *component.FaceDetectButton
	faceButtonBorder *canvas.Rectangle
	moveDetectButton *component.MoveDetectButton
	moveButtonBorder *canvas.Rectangle
	virtualCameraButton *widget.Button

	// Video dependencies
	recorder              recording.Recorder
	cam                   camera.Camera
	cameraWorker          *worker.CameraWorker
	frameProcessingWorker *worker.FrameProcessingWorker
	virtualCameraWorker   *worker.VirtualCameraWorker


	// Runtime state
	cancelContext context.CancelFunc
	isFaceDetect bool
	isMoveDetect bool
	chanMoveAnableDetect chan bool
	chanIsMoveDetect chan bool
	chanIsFaceDetect chan bool
	chanFaceMotionResultChan chan motiontracker.MotionResult

	// Prevent overlapping virtual-camera button operations.
	virtualCameraBusy bool

	// Callbacks
	onStopWindow func()
}

func NewVideoControlService(
    recorder recording.Recorder,
    cam camera.Camera,
    cameraWorker *worker.CameraWorker,
    frameProcessingWorker *worker.FrameProcessingWorker,
	virtualCameraWorker *worker.VirtualCameraWorker,
	onStopWindow func(),
) *VideoControlService {
	videoControlService :=  &VideoControlService{
		recorder:              recorder,
        cam:                   cam,
        cameraWorker:          cameraWorker,
        frameProcessingWorker: frameProcessingWorker,
		virtualCameraWorker:   virtualCameraWorker,
		chanIsFaceDetect: make(chan bool, 1),
		chanFaceMotionResultChan : make(chan motiontracker.MotionResult),
		chanIsMoveDetect: make(chan bool, 1), 
		chanMoveAnableDetect: make(chan bool, 1),
		onStopWindow: onStopWindow,
	}
	
	videoControlService.startButton = component.NewStartButton()
	videoControlService.startButton.SetOnTapped(videoControlService.start)

    videoControlService.stopButton = component.NewStopButton()
	videoControlService.stopButton.SetOnTapped(videoControlService.stop)

	//init border
	videoControlService.faceButtonBorder = canvas.NewRectangle(color.Transparent)
	videoControlService.faceDetectButton = component.NewFaceDetectButton()
	videoControlService.faceDetectButton.SetOnTapped(videoControlService.toggleFaceDetection)

	videoControlService.moveButtonBorder = canvas.NewRectangle(color.Transparent)
	videoControlService.moveDetectButton = component.NewMoveDetectButton()
	videoControlService.moveDetectButton.SetOnTapped(videoControlService.toggleMoveDetection)

    videoControlService.stopButton.Hide()

	// This button controls only virtual-camera output.
	videoControlService.virtualCameraButton = widget.NewButton(
		"Start Virtual Camera",
		videoControlService.toggleVirtualCamera,
	)
	videoControlService.virtualCameraButton.Disable()

	return videoControlService
	
}

func (s *VideoControlService) start() {
	
	log.Println("Starting camera...")

	s.startButton.Hide()
	s.stopButton.Show()
	s.virtualCameraButton.Enable()

	//TODO move to private method
	// Start FFmpeg.
	if !s.recorder.IsRunning() {
		if err := s.recorder.Start(); err != nil {

			log.Println(
				"failed to start recorder:",
				err,
			)

			//defer c.cam.Stop()

			return
		}
	}

	log.Println(
		"Recorder started",
	)
	/*TODO better to use 2 nestead context instead inside goroutine use timer := time.NewTimer(previewTimeout)

	// 1. Parent context managed by your Fyne UI "Stop" button
	parentCtx, cancelAll := context.WithCancel(context.Background())

	// 2. Inside your worker, create a child context with a timeout
	// If parentCtx is canceled by the button, childCtx cancels immediately!
	// If 5 seconds pass first, childCtx times out independently.
	childCtx, cancelTimeout := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancelTimeout() // Clean up timer resources when done
	*/
	var ctx context.Context

	ctx, s.cancelContext = context.WithCancel(
		context.Background(),
	)
		
	// 🟢 Motion detection listening
	s.motionHandler(ctx)

	// Start camera worker.
	//chanIsFaceDetect - chan event to handle state in goroutines
	
	frameChan, _/*faceImageChan*/ := s.cameraWorker.Run(ctx, s.chanIsFaceDetect)

	// Start preview and recording workers.
	//TODO maybe better - face motion goroutines move to frameProcessingWorker
	//TEST - At 30 FPS, you have approximately 33.33 ms per frame.
	//NEED TO Check if this place works not slower that 33.33ms - s.frameProcessingWorker
	s.frameProcessingWorker.Run(ctx, s.chanFaceMotionResultChan, s.chanIsMoveDetect, s.chanMoveAnableDetect)
	// Start the virtual-camera worker loop.
	// This loop consumes frames, but FFmpeg remains stopped until
	// the user clicks Start Virtual Camera.
	//TODO move to frameProcessingWorker or do not need 
	go s.virtualCameraWorker.Run(
		ctx,
		s.frameProcessingWorker.GetVirtualCameraFrame(),
	)
		
	// Receive frames from CameraWorker
	// and send them to FrameProcessingWorker.
	// this helps ti speed up - frame push to another channel in another goroutines and this helps to works faster
	// i try to delete this and pass frameChan directly to frameProcessingWorker.Run() and its works slower
	// push push values to ther goroutines inside frameProcessingWorker.Run()
	//TODO попробывать запустить этот метод в frameProcessingWorker.Run после goroutines + вместо 3-х каналов попробывать использовать 1
	go func() {
		s.frameProcessingWorker.PushJob(ctx, frameChan)		
	}()
}

func (s *VideoControlService) stop() {
	//TODO need to add channel to send bool value that user click on stop button - to add empty frame into preview
    log.Println("Stopping camera...")

	//TODO 
	/* for cancel logic better to add separate method 'close'
	// sync.Once guarantees the inner function runs only once, safely
	c.closeOnce.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
	})
	*/

	if s.cancelContext != nil {
		s.cancelContext()
	}

	// Stop the camera first so CameraWorker can finish.
	if !s.isMoveDetect {
		if err := s.cam.Stop(); err != nil {
			log.Println("camera stop:", err)
		}
	}

	// Wait until CameraWorker has completely stopped.
	s.cameraWorker.Wait()

	// Wait until frame processing has completely stopped.
	s.frameProcessingWorker.Wait()

	//instead recorder.Stop()
	/* check this
	// Stop recording after the processing workers have finished.
		if err := s.frameProcessingWorker.StopRecording(); err != nil {
			log.Printf("recorder stop: %v", err)
		}
	*/

	// Stop recorder after recording worker has finished.
	if err := s.recorder.Stop(); err != nil {
		log.Println("recorder stop:", err)
	}

	log.Println("Camera stopped")

	//stop fyne window with callback
	if s.onStopWindow != nil && !s.isMoveDetect {
		s.onStopWindow()
	}

	if s.isMoveDetect {
		s.stopButton.Hide()
		s.startButton.Show()

		//virtual cam logic
		s.virtualCameraButton.SetText("Start Virtual Camera")
		s.virtualCameraButton.Disable()
	}
}

func (s *VideoControlService) stopRecording() {
	log.Println("Stopping recorder...")

	// Wait until CameraWorker has completely stopped.
	s.cameraWorker.Wait()

	// Wait until frame processing has completely stopped.
	s.frameProcessingWorker.Wait()

	if s.recorder.IsRunning() {
		/*TODO recorder.Stop() replace with ->
		if err := s.frameProcessingWorker.StopRecording(); err != nil {
			log.Printf("recorder stop: %v", err)
		}
		*/
		if err := s.recorder.Stop(); err != nil {
			log.Println("recorder stop:", err)
		}
	}

	log.Println("Recorder stopped")
	//c.stopButton.Hide()
	//c.startButton.Show()
}

//listening for motion with logic to start recoring/stop if motion detected
func (s *VideoControlService) motionHandler(cancelContext context.Context) {
	go func(videoCtx context.Context) {
		log.Println("UI channel reader routine spawned")

		var holdTimer *time.Timer
		isGreenActive := false
		for {
			select {
			case <-videoCtx.Done():
				if holdTimer != nil {
					holdTimer.Stop()
				}

				log.Println("UI motion reader routine stopped")
				return

			case result, ok := <-s.chanFaceMotionResultChan:
				if !ok {
					return
				}

				// Only update the button design layers if face tracking toggle is enabled
				if s.isMoveDetect && result.IsMoving {
					//TODO if move detected - we can run capture process
					//todo and i think this logic need to move outside of the service

					//log.Println("move detected = ", result.IsMoving)
					
					select {
					case s.chanIsMoveDetect <- true:
					default:
						// Remove the previous unread state.
						<-s.chanIsMoveDetect

						// Send the newest state.
						s.chanIsMoveDetect <- true
					}


					// If a timer is already running, stop it to extend the green time
					if holdTimer != nil {
						holdTimer.Stop()
					}

					// Turn button background green if it isn't already
					if !isGreenActive {
						isGreenActive = true
						fyne.Do(func() {
							moveDetectButton := s.moveDetectButton.GetButton()
							moveDetectButton.Importance = widget.HighImportance
							moveDetectButton.Refresh()
							s.moveDetectButton.SetButton(moveDetectButton)
						})
					}

					log.Println("move detect")

					if !s.recorder.IsRunning() {
						// Start FFmpeg.
						if err := s.recorder.Start(); err != nil {

							log.Println(
								"failed to start recorder:",
								err,
							)

							defer s.cam.Stop()

							return
						}
					}

					// Start a new 20-second timer to reset the color back to normal and stop recording
					holdTimer = time.AfterFunc(20*time.Second, func() {
						// Ensure context isn't closed before resetting UI
						select {
						case <-videoCtx.Done():
							return
						default:
							isGreenActive = false
							fyne.Do(func() {
								moveDetectButton := s.moveDetectButton.GetButton()
								moveDetectButton.Importance = widget.MediumImportance
								moveDetectButton.Refresh()
								s.moveDetectButton.SetButton(moveDetectButton)
							})
							log.Println("[UI] 20-second hold finished. Resetting button color to default.")
							//and now stop recording after 20 seconds

							select {
							case s.chanIsMoveDetect <- false:
							default:
								// Remove the previous unread state.
								<-s.chanIsMoveDetect

								// Send the newest state.
								s.chanIsMoveDetect <- false
							}

							//stop all record 
							//TODO need to do new function to stop only record and do not close the script
							//TODO need to add new function to stop only record and do not stop another logic
							s.stopRecording()
							//c.stop()
						}
					})
				}
			}
		}
	}(cancelContext)
}

func (s *VideoControlService) toggleFaceDetection() {
	s.isFaceDetect = !s.isFaceDetect

	if s.isFaceDetect {
		
		s.faceButtonBorder.StrokeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
		s.faceButtonBorder.StrokeWidth = 3
		//faceDetectButton.SetText("State: ACTIVE")
	} else {
		s.faceButtonBorder.StrokeColor = color.Transparent
		s.faceButtonBorder.StrokeWidth = 0
		//faceDetectButton.SetText("Toggle State")
	}

	s.faceButtonBorder.Refresh()

	// Send the latest state without blocking the UI.
	select {
	case s.chanIsFaceDetect <- s.isFaceDetect:
	default:
		// Remove the previous unread state.
		<-s.chanIsFaceDetect

		// Send the newest state.
		s.chanIsFaceDetect <- s.isFaceDetect
	}
}

func (s *VideoControlService) toggleMoveDetection() {
	s.isMoveDetect = !s.isMoveDetect

	if s.isMoveDetect {
		s.moveButtonBorder.StrokeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
		s.moveButtonBorder.StrokeWidth = 3

		//run preview and video record
		s.start()
	} else {
		s.moveButtonBorder.StrokeColor = color.Transparent
		s.moveButtonBorder.StrokeWidth = 0
	}

	s.moveButtonBorder.Refresh()

	select {
	case s.chanMoveAnableDetect <- s.isMoveDetect:
	default:
		// Remove the previous unread state.
		<-s.chanMoveAnableDetect

		// Send the newest state.
		s.chanMoveAnableDetect <- s.isMoveDetect
	}
}

// toggleVirtualCamera toggles only virtual-camera output.
// toggleVirtualCamera toggles only virtual-camera output.
// The normal camera pipeline continues regardless of this button's state.
func (s *VideoControlService) toggleVirtualCamera() {
	/*
	s.stateMu.Lock()

	if !s.isStarted || s.isStopping || s.virtualCameraBusy {
		s.stateMu.Unlock()
		return
	}*/

	s.virtualCameraBusy = true
	//s.virtualCameraButton.Disable()

	//s.stateMu.Unlock()

	if s.virtualCameraWorker.Enabled() {
		go func() {
			err := s.virtualCameraWorker.StopOutput()

			if err != nil {
				log.Printf("stop virtual camera: %v", err)
			}

			fyne.Do(func() {

				//s.stateMu.Lock()
				s.virtualCameraBusy = false
				//isStarted := s.isStarted && !s.isStopping
				//s.stateMu.Unlock()

				s.virtualCameraButton.SetText("Start Virtual Camera")

				//if isStarted {
				s.virtualCameraButton.Enable()
				//}
			})
		}()

		return
	}

	go func() {
		//TODO this logic i think better to move into worker
		err := s.virtualCameraWorker.StartOutput()

		if err != nil {
			log.Printf("start virtual camera: %v", err)
		}

		fyne.Do(func() {
			//s.stateMu.Lock()
			s.virtualCameraBusy = false
			//isStarted := s.isStarted && !s.isStopping
			//s.stateMu.Unlock()

			if err != nil {
				s.virtualCameraButton.SetText("Start Virtual Camera")
			} else {
				s.virtualCameraButton.SetText("Stop Virtual Camera")
			}

			//if isStarted {
			s.virtualCameraButton.Enable()
			//}
		})
	}()
}

func (c *VideoControlService) Buttons() fyne.CanvasObject {
    faceDetectButtonContainer := container.NewStack(
		c.faceButtonBorder,
		c.faceDetectButton.GetButton(),
	)
	moveDetectButtonContainer := container.NewStack(
		c.moveButtonBorder,
		c.moveDetectButton.GetButton(),
	)

	return container.NewHBox(
		c.startButton.GetButton(),
		c.stopButton.GetButton(),
		c.virtualCameraButton,
		faceDetectButtonContainer,
		moveDetectButtonContainer,
	)
}
