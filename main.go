package main

import (
	"log"
	"time"
	//"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"video/internal/camera"
	"video/internal/preview"
	"video/internal/recording"
	"video/internal/detection"
)

const (
	cameraDevice = "/dev/video0"

	// Camera configuration.
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

	// ----------------------------------------
	// State
	// ----------------------------------------
	//var mu sync.Mutex
	var running bool
	var stopChan chan struct{}

	// ----------------------------------------
	// Start / Stop button
	// ----------------------------------------
	var startStopButton *widget.Button
	startStopButton = widget.NewButton(
		"Start Video & Recording",
		func() {
			//mu.Lock()

			if running {
				//mu.Unlock()
				// =====================================
				// STOP
				// =====================================

				log.Println(
					"Stopping camera...",
				)

				close(stopChan)

				_ = cam.Stop()

				if err := recorder.Stop(); err != nil {
					log.Println(
						"failed to stop recorder:",
						err,
					)
				}

				running = false

				startStopButton.SetText(
					"Start Video & Recording",
				)

				startStopButton.Importance =
					widget.MediumImportance

				startStopButton.Refresh()

				//close windget window
				myWindow.Close()

				log.Println(
					"Camera stopped",
				)

				//if delete return - script can continue to work
				return
			}

			// =====================================
			// START
			// =====================================


			log.Println("Starting camera...")

			if err := cam.Start(); err != nil {
				//mu.Unlock()

				log.Println(
					"failed to start camera:",
					err,
				)

				return
			}

			log.Println(
				"Camera started",
			)

			// Start FFmpeg.
			if err := recorder.Start(); err != nil {

				log.Println(
					"failed to start recorder:",
					err,
				)

				_ = cam.Stop()

				//mu.Unlock()

				return
			}

			log.Println(
				"Recorder started",
			)

			stopChan = make(chan struct{})

			running = true

			//mu.Unlock()

			// Start frame processing.
			go processFrames(
				cam,
				detector,
				videoPreview,
				recorder,
				stopChan,
			)

			startStopButton.SetText(
				"Stop Video & Recording",
			)

			startStopButton.Importance = widget.DangerImportance
			startStopButton.Refresh()

			return
		},
	)


	// ----------------------------------------
	// Buttons
	// ----------------------------------------

	//buttons := container.NewHBox(
	//	startStopButton,
	//	closeButton,
	//)

	// ----------------------------------------
	// Layout
	// ----------------------------------------

	content := container.NewBorder(
		startStopButton,//buttons,
		nil,
		nil,
		nil,
		videoPreview.Widget(),
	)

	myWindow.SetContent(content)

	// ----------------------------------------
	// Run
	// ----------------------------------------

	myWindow.ShowAndRun()
}

// processFrames continuously reads frames
// from the camera and sends each frame to:
//
// 1. Fyne preview
// 2. FFmpeg recorder
func processFrames(
	cam camera.Camera,
	detector detection.Detector,
	preview *preview.FynePreview,
	recorder recording.Recorder,
	stopChan <-chan struct{},
) {

	log.Println(
		"Frame processing started",
	)

	frameNumber := 0

	var lastFaces []detection.Face

	for {

		// Check stop signal.
		select {

		case <-stopChan:

			log.Println(
				"Frame processing stopped",
			)

			return

		default:
		}

		// ------------------------------------
		// Read ONE frame
		// ------------------------------------

		frame, err := cam.Read()

		if err != nil {
			log.Println(
				"camera read:",
				err,
			)

			continue
		}

		// ====================================
		// Detect faces
		// ====================================
	
		if frameNumber%5 == 0 {

			faces, err := detector.Detect(frame)

			if err != nil {
				log.Println("face detection:", err)
			} else {
				// Save the latest detection.
				lastFaces = faces
			}
		}

		// Draw the last detected face.
		frameWithFaces, err := detector.DrawFaces(
			frame,
			lastFaces,
		)
		if err != nil {
			log.Println("draw faces:", err)
			continue
		}

		/*
		faces, err := detector.Detect(
			frame,
		)

		if err != nil {

			log.Println(
				"face detection:",
				err,
			)

			continue
		}*/

		// ------------------------------------
		// Send frame to Fyne - preview
		// ------------------------------------

		if err := preview.ShowFrame(frameWithFaces); err != nil {
			log.Println(
				"preview:",
				err,
			)
		}

		// ------------------------------------
		// Send frame to FFmpeg - to save
		// ------------------------------------

		if err := recorder.WriteFrame(frameWithFaces); err != nil {
			log.Println(
				"recorder:",
				err,
			)
		}

		//iterate frames
		frameNumber++
	}
}
