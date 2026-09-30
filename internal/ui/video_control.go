package ui

import (
	"context"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/facemove"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/worker"
)

type VideoControl struct {
    // UI
	startButton      *widget.Button
	stopButton       *widget.Button
	faceDetectButton *widget.Button
	faceButtonBorder *canvas.Rectangle
	moveDetectButton *widget.Button
	moveButtonBorder *canvas.Rectangle

	// Video dependencies
	recorder              recording.Recorder
	cam                   camera.Camera
	cameraWorker          *worker.CameraWorker
	frameProcessingWorker *worker.FrameProcessingWorker

	// Runtime state
	ctx            context.Context
	cancel         context.CancelFunc
	isFaceDetect   bool
	isMoveDetect   bool
	chanIsFaceDetect chan bool
	chanFaceMotionResultChan chan facemove.MotionResult

	// Callbacks
	onStopWindow func()
}

func NewVideoControl(
    recorder recording.Recorder,
    cam camera.Camera,
    cameraWorker *worker.CameraWorker,
    frameProcessingWorker *worker.FrameProcessingWorker,
	onStopWindow func(),
) *VideoControl {

    controls := &VideoControl{
        recorder:              recorder,
        cam:                   cam,
        cameraWorker:          cameraWorker,
        frameProcessingWorker: frameProcessingWorker,
		chanIsFaceDetect: make(chan bool, 1),
		chanFaceMotionResultChan : make(chan facemove.MotionResult),
		onStopWindow: onStopWindow,
    }

    controls.startButton = widget.NewButton(
        "Start Video",
        controls.start,
    )

    controls.stopButton = widget.NewButton(
        "Stop Video",
        controls.stop,
    )

	//init border
	controls.faceButtonBorder = canvas.NewRectangle(color.Transparent)
	controls.faceDetectButton = widget.NewButton(
		"Face detect",
		controls.toggleFaceDetection,
	)

	controls.moveButtonBorder = canvas.NewRectangle(color.Transparent)
	controls.moveDetectButton = widget.NewButton(
		"Move detect",
		controls.toggleMoveDetection,
	)



    controls.stopButton.Hide()

    return controls
}

func (c *VideoControl) start() {
	/*
    log.Println("Starting camera...")

    if err := c.recorder.Start(); err != nil {
        log.Println("failed to start recorder:", err)
        return
    }

    c.ctx, c.cancel = context.WithCancel(context.Background())

    c.startButton.Hide()
    c.stopButton.Show()

    frameChan, _ := c.cameraWorker.Run(
        c.ctx,
        // face detection channel
    )

    c.frameProcessingWorker.Run(c.ctx)

    go func() {
        c.frameProcessingWorker.PushJob(
            c.ctx,
            frameChan,
        )
    }()*/
	log.Println("Starting camera...")

	c.startButton.Hide()
	c.stopButton.Show()

	// Start FFmpeg.
	if err := c.recorder.Start(); err != nil {

		log.Println(
			"failed to start recorder:",
			err,
		)

		defer c.cam.Stop()
		//_ = cam.Stop()


		return
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

	c.ctx, c.cancel = context.WithCancel(
		context.Background(),
	)
		
	// 🟢 FACE Motion detection listening
	// This goroutine runs instantly when the video goes live. It keeps the channel 
	// drained so the worker never flags a "channel full" frame drop.
	// listeting goroutinmes to check if face motion detection
	
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

			case result, ok := <-c.chanFaceMotionResultChan:
				if !ok {
					return
				}

				// Only update the button design layers if face tracking toggle is enabled
				if c.isMoveDetect && result.IsMoving {
					//TODO if move detected - we can run capture process
					//todo and i think this logic need to move outside of the service

					log.Println("move detected = ", result.IsMoving)


					// If a timer is already running, stop it to extend the green time
					if holdTimer != nil {
						holdTimer.Stop()
					}

					// Turn button background green if it isn't already
					if !isGreenActive {
						isGreenActive = true
						fyne.Do(func() {
							c.moveDetectButton.Importance = widget.HighImportance
							c.moveDetectButton.Refresh()
						})
					}

					// Start a new 5-second timer to reset the color back to normal
					holdTimer = time.AfterFunc(5*time.Second, func() {
						// Ensure context isn't closed before resetting UI
						select {
						case <-videoCtx.Done():
							return
						default:
							isGreenActive = false
							fyne.Do(func() {
								c.moveDetectButton.Importance = widget.MediumImportance
								c.moveDetectButton.Refresh()
							})
							log.Println("[UI] 5-second hold finished. Resetting button color to default.")
						}
					})
				}
			}
		}
	}(c.ctx)


	// Start camera worker.
	//chanIsFaceDetect - chan event to handle state in goroutines
	frameChan, _/*faceImageChan*/ := c.cameraWorker.Run(c.ctx, c.chanIsFaceDetect)

	 // Start preview and recording workers.
	c.frameProcessingWorker.Run(c.ctx, c.chanFaceMotionResultChan)

			
	// Receive frames from CameraWorker
	// and send them to FrameProcessingWorker.
	go func() {
		c.frameProcessingWorker.PushJob(c.ctx, frameChan/*, faceImage*/)		
	}()
}

func (c *VideoControl) stop() {
    log.Println("Stopping camera...")

	if c.cancel != nil {
		c.cancel()
	}

	// Stop the camera first so CameraWorker can finish.
	if err := c.cam.Stop(); err != nil {
		log.Println("camera stop:", err)
	}

	// Wait until CameraWorker has completely stopped.
	c.cameraWorker.Wait()

	// Wait until frame processing has completely stopped.
	c.frameProcessingWorker.Wait()

	// Stop recorder after recording worker has finished.
	if err := c.recorder.Stop(); err != nil {
		log.Println("recorder stop:", err)
	}

	log.Println("Camera stopped")

	//stop fyne window with callback
	if c.onStopWindow != nil {
		c.onStopWindow()
	}

	
}

func (c *VideoControl) toggleFaceDetection() {
	c.isFaceDetect = !c.isFaceDetect

	if c.isFaceDetect {
		
		c.faceButtonBorder.StrokeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
		c.faceButtonBorder.StrokeWidth = 3
		//faceDetectButton.SetText("State: ACTIVE")
	} else {
		c.faceButtonBorder.StrokeColor = color.Transparent
		c.faceButtonBorder.StrokeWidth = 0
		//faceDetectButton.SetText("Toggle State")
	}

	c.faceButtonBorder.Refresh()

	// Send the latest state without blocking the UI.
	select {
	case c.chanIsFaceDetect <- c.isFaceDetect:
	default:
		// Remove the previous unread state.
		<-c.chanIsFaceDetect

		// Send the newest state.
		c.chanIsFaceDetect <- c.isFaceDetect
	}
}

func (c *VideoControl) toggleMoveDetection() {
	c.isMoveDetect = !c.isMoveDetect

	if c.isMoveDetect {
		
		c.moveButtonBorder.StrokeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
		c.moveButtonBorder.StrokeWidth = 3
	} else {
		c.moveButtonBorder.StrokeColor = color.Transparent
		c.moveButtonBorder.StrokeWidth = 0
	}

	c.moveButtonBorder.Refresh()

	// Send the latest state without blocking the UI.
	/*
	select {
	case c.chanIsFaceDetect <- c.isMoveDetect:
	default:
		// Remove the previous unread state.
		<-c.chanIsFaceDetect

		// Send the newest state.
		c.chanIsFaceDetect <- c.isMoveDetect
	}*/
}

func (c *VideoControl) Buttons() fyne.CanvasObject {
    faceDetectButtonContainer := container.NewStack(
		c.faceButtonBorder,
		c.faceDetectButton,
	)
	moveDetectButtonContainer := container.NewStack(
		c.moveButtonBorder,
		c.moveDetectButton,
	)

	return container.NewHBox(
		c.startButton,
		c.stopButton,
		faceDetectButtonContainer,
		moveDetectButtonContainer,
	)
}