package main

import (
	"log"
	"time"

	"bytes"
	"encoding/xml"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/detection"
	"github.com/bogdanCap/video-cap/internal/motiontracker"
	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/ui"
	"github.com/bogdanCap/video-cap/internal/worker"
	"github.com/bogdanCap/video-cap/internal/virtualcamera"
)

const (
	//cameraDevice = "/dev/video1"
	cameraDevice = "/dev/video0"
	virtualCameraDevice = "/dev/video2"

	// Camera configuration.
	//SD (480p): 640 × 480
	//           800 × 600
	//           1024 × 768
	//HD (720p): 1280 × 720
	//Full HD (1080p): 1920 × 1080
	//QHD / 2K (1440p): 2560 × 1440
	cameraWidth  = 640
	cameraHeight = 480
	//detectionWidth = 640
	//detectionHeight = 360
	cameraFPS    = 30

	// Saved video configuration.
	recordingWidth  = 1920
	recordingHeight = 1080

	// Directory for MP4 files.
	recordingDir = "recordings"

	// Each file is 30 seconds.
	chunkDuration = 30 * time.Second

	//cascadePath = "model/haarcascade_frontalface_default.xml"
	cascadePath = "model/pigo_facefinder"

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
	//TODO for face detector better to use v4l + pigo
	detector, err := detection.NewFaceDetector(
		cascadePath,
	)

	if err != nil {
		log.Fatal(err)
	}

	// ----------------------------------------
	// Preview
	// ----------------------------------------

	videoPreview := preview.NewFynePreview(
		cameraWidth,//320,//cameraWidth,
		cameraHeight,//240,//cameraHeight,
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

	motionTracker := motiontracker.NewMotionTracker(
		5.0,
	)

	// Frame processing.
	frameProcessingWorker := worker.NewFrameProcessingWorker(
		videoPreview,
		recorder,
		motionTracker,
	)

	virtualCamera := virtualcamera.NewFFmpegVirtualCamera(
		virtualCameraDevice,
		cameraWidth,
		cameraHeight,
		cameraFPS,
	)

	virtualCameraWorker := worker.NewVirtualCameraWorker(virtualCamera)

	/** TODO init new logic related with face saving 
		ctx := context.Background()

		// PostgreSQL.
		db, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/mydb")
		if err != nil {
			panic(err)
		}
		defer db.Close()

		// Face detection.
		detector := detection.NewFaceDetector()

		// need to init befor ArcFace
		ort.SetSharedLibraryPath(
			"/usr/local/lib/libonnxruntime.so",
		)

		if err := ort.InitializeEnvironment(); err != nil {
			panic(err)
		}

		defer ort.DestroyEnvironment()

		// Face recognition.
		recognizer, err := recognition.NewArcFace(
			"model/arcface.onnx",
			"input",
			"output",
		)

		if err != nil {
			panic(err)
		}
		defer recognizer.Close()

		// Repository.
		faceRepository := repository.NewFaceRepository(db)

		// Business service.
		faceService := face.NewFaceService(
			detector,
			recognizer,
			faceRepository,
		)

		_ = faceService
	*/

	// ----------------------------------------
	// Fyne
	// ----------------------------------------

	uiService := ui.NewUIService(
		recorder, 
		cam, 
		virtualCameraWorker,
		frameProcessingWorker, 
		cameraWorker, 
		videoPreview,
	)
	uiService.Start()
}

func xmlEscape(s string) string {
	var buf bytes.Buffer
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return s
	}
	return buf.String()
}