package main

import (
	"log"
	"time"
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/detection"
	"github.com/bogdanCap/video-cap/internal/worker"
)

const (
	cameraDevice = "/dev/video0"

	// Camera configuration.
	//cameraWidth = 2560
	//cameraHeight = 1440
	cameraWidth  = 1920
	cameraHeight = 1080
	cameraFPS    = 30

	// Saved video configuration.
	recordingWidth  = 1920
	recordingHeight = 1080

	// Directory for MP4 files.
	recordingDir = "recordings"

	// Each file is 30 seconds.
	chunkDuration = 30 * time.Second

	cascadePath = "model/haarcascade_frontalface_default.xml"

	//frameBufferSize = 2
)

func main() {

	// ----------------------------------------
	// Camera
	// ----------------------------------------

	cam, err := camera.NewV4LCamera(
		cameraDevice,
		cameraWidth,
		cameraHeight,
		cameraFPS,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer cam.Close()

	// ========================================
	// Face detector
	// ========================================

	detector, err := detection.NewFaceDetector(
		cascadePath,
		cameraWidth,
		cameraHeight,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer detector.Close()

	// ----------------------------------------
	// Preview
	// ----------------------------------------

	videoPreview := preview.NewFynePreview(
		//cameraWidth,
		//cameraHeight,
	)

	// ----------------------------------------
	// Recorder
	// ----------------------------------------

	recorder := recording.NewFFmpegRecorder(
		recordingDir,
		recordingWidth,
		recordingHeight,
		cameraFPS,
		chunkDuration,
	)

	// Camera worker. - read frame by frame from camera
	cameraWorker := worker.NewCameraWorker(
		cam,
		detector,
	)

	// Frame processing.
	frameProcessingWorker := worker.NewFrameProcessingWorker(
		videoPreview,
		recorder,
	)


	// ----------------------------------------
	// Fyne
	// ----------------------------------------

	myApp := app.New()

	myWindow := myApp.NewWindow(
		"USB Camera",
	)

	myWindow.Resize(
		fyne.NewSize(800, 500),
	)


	var (
		ctx context.Context
		cancel context.CancelFunc
		//running bool
		//processingWG sync.WaitGroup
		//frameWG sync.WaitGroup
	)

	// ----------------------------------------
	// Start
	// ----------------------------------------
	startButton := widget.NewButton(
		"Start Video",
		func() {
			log.Println("Starting camera...")


			// Start FFmpeg.
			if err := recorder.Start(); err != nil {

				log.Println(
					"failed to start recorder:",
					err,
				)

				_ = cam.Stop()


				return
			}

			log.Println(
				"Recorder started",
			)

			ctx, cancel = context.WithCancel(
				context.Background(),
			)

			// Start preview and recording workers.
			frameProcessingWorker.Run(ctx)

			// Start camera worker.
			frameChan := cameraWorker.Run(ctx)

			// Receive frames from CameraWorker
			// and send them to FrameProcessingWorker.
			go func() {

				for {
					select {
					case <-ctx.Done():
						return

					case frame, ok := <-frameChan:
						if !ok {
							return
						}

						frameProcessingWorker.Process(ctx, frame)
					}
				}
			}()
		},
	)

	stopButton := widget.NewButton(
		"Stop Video",
		func() {
			log.Println(
				"Stopping camera...",
			)


			if cancel != nil {
				cancel()
			}

			_ = cam.Stop()

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


	// ----------------------------------------
	// Layout
	// ----------------------------------------

	buttons := container.NewHBox(startButton, stopButton)

	content := container.NewBorder(
		nil,
		buttons,
		nil,
		nil,
		//nil,
		videoPreview.Widget(),
	)

	myWindow.SetContent(content)

	/*
	myWindow.Resize( 
		fyne.NewSize(800, 500), 
	)
	*/
	// ----------------------------------------
	// Run
	// ----------------------------------------

	myWindow.ShowAndRun()
}
