package main

import (
	"log"
	"time"
	//"sync"
	//"image"

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

	frameBufferSize = 2
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

	// --------------------------------------------------
	// Channels
	// --------------------------------------------------

	previewChan := make(chan []byte, frameBufferSize)
	recordingChan := make(chan []byte, frameBufferSize)

	// --------------------------------------------------
	// Start processing consumers (listening face detection and recording channel)
	// --------------------------------------------------
	startProcessing(
		previewChan,
		recordingChan,
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


				return
			}

			log.Println(
				"Recorder started",
			)

			//stopChan = make(chan struct{})
			running = true
			//faceDetectionEnabled := true

			// Create stop channel for this session.
			stopChan = make(chan struct{})

			//mu.Unlock()

			// Start frame processing.
			go runCamera(
				cam,
				detector,
				previewChan,
				recordingChan,
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


// runCamera is the producer.
//
// It:
//   - reads frames from the camera
//   - processes the frame
//   - sends the processed frame to preview
//   - sends the processed frame to recording
//
// There are no goroutines here.
func runCamera(
	cam camera.Camera,
	detector detection.Detector,
	previewChan chan<- []byte,
	recordingChan chan<- []byte,
	stopChan <-chan struct{},
) {
	var frameNumber int

	var lastFaces []detection.Face

	for {
		// ----------------------------------------
		// Check stop signal to stop recording
		// ----------------------------------------

		select {
		case <-stopChan:
			return

		default:
		}

		// --------------------------------------------------
		// Read frame
		// --------------------------------------------------

		frame, err := cam.Read()
		if err != nil {
			log.Println("camera read:", err)

			return
		}

		frameNumber++

		// --------------------------------------------------
		// Process frame
		// --------------------------------------------------

		frame, lastFaces = processFrames(
			frame,
			frameNumber,
			lastFaces,
			detector,
			true,
		)

		// --------------------------------------------------
		// Send frame to consumers
		// --------------------------------------------------

		//previewChan <- frame
		//recordingChan <- frame
		// ----------------------------------------
		// Preview
		// ----------------------------------------

		// Preview must never block camera.
		//
		// If Fyne is busy, simply drop this frame.
		select {

		case previewChan <- frame:

		default:
			// Preview is busy.
			// Drop frame.
		}

		// ----------------------------------------
		// Recording
		// ----------------------------------------

		// Recording should not randomly drop frames.
		select {

		case recordingChan <- frame:

		case <-stopChan:
			return
		}
	}
}



// processFrames contains all face detection
// and drawing logic.
func processFrames(
	frame []byte,
	frameNumber int,
	lastFaces []detection.Face,
	detector detection.Detector,
	faceDetectionEnabled bool,
) ([]byte, []detection.Face) {

	if !faceDetectionEnabled {
		return frame, nil
	}

	// --------------------------------------------------
	// Detect face every 5th frame
	// --------------------------------------------------

	if frameNumber%5 == 0 {
		faces, err := detector.Detect(frame)
		if err != nil {
			log.Println("face detection:", err)

			return frame, lastFaces
		}

		// We only need one face.
		if len(faces) > 0 {
			lastFaces = faces[:1]
			// Remember the new rectangle.
			//face := faces[0]
			//lastFace = &face
		} else {
			lastFaces = nil
		}
	}

	// --------------------------------------------------
	// Draw face
	// --------------------------------------------------

	//if lastFace != nil {
	if len(lastFaces) > 0 {
		processedFrame, err := detector.DrawFaces(
			frame,
			lastFaces,
		)
		//save last detection rectangles
		//processedFrame, err := detector.DrawFaces(
		//	frame,
		//	[]detection.Face{*lastFace},
		//)
		if err != nil {
			log.Println("draw faces:", err)

			return frame, lastFaces
		}

		frame = processedFrame
	}

	return frame, lastFaces
}


func startProcessing(
	previewChan <-chan []byte,
	recordingChan <-chan []byte,
	videoPreview *preview.FynePreview,
	recorder recording.Recorder,
) {
	// --------------------------------------------------
	// Preview consumer
	// --------------------------------------------------

	go func() {
		for frame := range previewChan {
			if err := videoPreview.ShowFrame(frame); err != nil {
				log.Println("preview:", err)
			}
		}
	}()

	// --------------------------------------------------
	// Recording consumer
	// --------------------------------------------------

	go func() {
		for frame := range recordingChan {
			if err := recorder.WriteFrame(frame); err != nil {
				log.Println("recording:", err)
			}
		}
	}()
}
