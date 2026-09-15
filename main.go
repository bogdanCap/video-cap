package main

import (
	"log"
	"time"

	"github.com/bogdanCap/video-cap/internal/ui"
	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/detection"
	"github.com/bogdanCap/video-cap/internal/worker"
)

const (
	cameraDevice = "/dev/video0"

	// Camera configuration.
	//SD (480p): 640 × 480
	//HD (720p): 1280 × 720
	//Full HD (1080p): 1920 × 1080
	//QHD / 2K (1440p): 2560 × 1440
	cameraWidth  = 640
	cameraHeight = 480
	detectionWidth = 640
	detectionHeight = 360
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
		detectionWidth,
		detectionHeight,
	)

	if err != nil {
		log.Fatal(err)
	}

	defer detector.Close()

	// ----------------------------------------
	// Preview
	// ----------------------------------------

	videoPreview := preview.NewFynePreview(
		cameraWidth,
		cameraHeight,
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

	uiService := ui.NewUIService(
		recorder, 
		cam, 
		frameProcessingWorker, 
		cameraWorker, 
		videoPreview,
	)
	uiService.Start()
}
