package ui

import (
	"log"
	//"time"
	"context"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	//"github.com/bogdanCap/video-cap/internal/facemove"
	"github.com/bogdanCap/video-cap/internal/worker"

)

type WindowService struct {
	App    fyne.App
	Window fyne.Window
}

func NewUIService(
	recorder recording.Recorder, 
	cam camera.Camera,
	frameProcessingWorker *worker.FrameProcessingWorker,
	cameraWorker *worker.CameraWorker,
	videoPreview *preview.FynePreview,
) *WindowService {
	myApp := app.New()

	myWindow := myApp.NewWindow(
		"USB Camera",
	)


	var (
		ctx context.Context
		cancel context.CancelFunc
		startButton *widget.Button
		stopButton *widget.Button
		faceDetectButton *widget.Button
		isFaceDetect bool
	)
	//channel with event to toggle face detection logic
	chanIsFaceDetect := make(chan bool, 1)
	//chanFaceMotionResultChan := make(chan facemove.MotionResult, 1)

	// Custom canvas background rectangle for changing face detect button colors
	faceButtonBorder := canvas.NewRectangle(color.Transparent)
	faceButtonBorder.StrokeWidth = 0

	// ----------------------------------------
	// Start
	// ----------------------------------------
	startButton = widget.NewButton(
		"Start Video",
		func() {
			log.Println("Starting camera...")

			startButton.Hide()
			stopButton.Show()

			// Start FFmpeg.
			if err := recorder.Start(); err != nil {

				log.Println(
					"failed to start recorder:",
					err,
				)

				defer cam.Stop()
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

			ctx, cancel = context.WithCancel(
				context.Background(),
			)

			

			// 🟢 FACE Motion detection listening
			// This goroutine runs instantly when the video goes live. It keeps the channel 
			// drained so the worker never flags a "channel full" frame drop.
			// listeting goroutinmes to check if face motion detection
			/*
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

					case result, ok := <-chanFaceMotionResultChan:
						if !ok {
							return
						}

						// Only update the button design layers if face tracking toggle is enabled
						if isFaceDetect && result.IsMoving {
							
							// If a timer is already running, stop it to extend the green time
							if holdTimer != nil {
								holdTimer.Stop()
							}

							// Turn button background green if it isn't already
							if !isGreenActive {
								isGreenActive = true
								fyne.Do(func() {
									faceDetectButton.Importance = widget.HighImportance
									faceDetectButton.Refresh()
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
										faceDetectButton.Importance = widget.MediumImportance
										faceDetectButton.Refresh()
									})
									log.Println("[UI] 5-second hold finished. Resetting button color to default.")
								}
							})
						}
					}
				}
			}(ctx)
			*/


			// Start camera worker.
			//chanIsFaceDetect - chan event to handle state in goroutines
			frameChan, _/*faceImageChan*/ := cameraWorker.Run(ctx, chanIsFaceDetect)

			// Start preview and recording workers.
			frameProcessingWorker.Run(ctx/*, chanFaceMotionResultChan*/)

			
			// Receive frames from CameraWorker
			// and send them to FrameProcessingWorker.
			go func() {
				frameProcessingWorker.PushJob(ctx, frameChan/*, faceImage*/)
				
				/*
				for {
					select {
					case <-ctx.Done():
						log.Println("frame distributor stopped")
						return

					case frame, ok := <-frameChan:
						if !ok {
							return
						}

						//faceImage := <-faceImageChan

						frameProcessingWorker.PushJob(ctx, frame)
					}
				}*/
			}()
		},
	)

	stopButton = widget.NewButton(
		"Stop Video",
		func() {
			log.Println("Stopping camera...")

			if cancel != nil {
				cancel()
			}

			defer cam.Stop()
			//_ = cam.Stop()

			//todo wait when worker finished
			// Wait for CameraWorker.
			cameraWorker.Wait()

			// FrameProcessingWorker now:
			//
			// - stops preview
			// - drains recordingChan
			// - waits for all WriteFrame() goroutines
			frameProcessingWorker.Wait()

			if err := recorder.Stop(); err != nil {
				log.Println("recorder stop:", err)
			}

			//running = false

			log.Println("Camera stopped")

			myWindow.Close()
		},
	)
	//by default do not need to show
	stopButton.Hide()

	//faceButtonBorder := canvas.NewRectangle(color.Transparent)
	//faceButtonBorder.StrokeWidth = 0

	faceDetectButton = widget.NewButton("Face detect", func() {
		// Change the boolean value to its opposite
		isFaceDetect = !isFaceDetect

		if isFaceDetect {
			// Change border to green and give it a thickness
			faceButtonBorder.StrokeColor = color.RGBA{R: 0, G: 255, B: 0, A: 255} // Green
			faceButtonBorder.StrokeWidth = 3
			//faceDetectButton.SetText("State: ACTIVE")

			
			
		} else {
			// Revert border to transparent
			faceButtonBorder.StrokeColor = color.Transparent
			faceButtonBorder.StrokeWidth = 0
			//faceDetectButton.SetText("Toggle State")
		}

		// Refresh the border canvas object to reflect structural updates
		faceButtonBorder.Refresh()

		// 3. Send the updated value to the channel without blocking the UI.
		// We use a select with a default case so if the channel buffer is full,
		// it drains the old value and sends the newest state.
		select {
		case chanIsFaceDetect <- isFaceDetect:
		default:
			<-chanIsFaceDetect // Remove old unread state
			chanIsFaceDetect <- isFaceDetect
		}
	})


	// ----------------------------------------
	// Layout
	// ----------------------------------------

	faceDetectButtonContainer := container.NewStack(faceButtonBorder, faceDetectButton)
	buttons := container.NewHBox(startButton, stopButton, faceDetectButtonContainer)

	content := container.NewBorder(
		nil,
		buttons,
		nil,
		nil,
		//nil,
		videoPreview.Widget(),
	)
	//content := container.NewPadded(buttons)

	myWindow.SetContent(content)

	//800/480 rp5 disaply resolution
	myWindow.Resize(
		fyne.NewSize(800, 480),
	)

	// ----------------------------------------
	// Run
	// ----------------------------------------

	//myWindow.ShowAndRun()

	return &WindowService{
		App:    myApp,
		Window: myWindow,
	}
}

func (s *WindowService) Start() {
	s.Window.ShowAndRun()
}
