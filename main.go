package main

import (
	"log"
	"time"
	"os"
	"embed"
	"context"
	"strconv"
	
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

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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
	recordingWidth  = 1024//1024 //1920
	recordingHeight = 768//720 //1080

	// Directory for MP4 files.
	recordingDir = "recordings"

	// Each file is 30 seconds.
	chunkDuration = 30 * time.Second

	//cascadePath = "model/haarcascade_frontalface_default.xml"
	cascadePath = "model/pigo_facefinder"

	//frameBufferSize = 2
)

var embedMigrations embed.FS

func main() {
	//TODO move to db connection service
	ctx := context.Background()

	// Load .env file if present (ignore error if running with system env vars)
	_ = godotenv.Load()

	/*
	dbURL := os.Getenv("DATABASE_URL")
	// 2. Parse the config and/or pool settings
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE_URL: %v", err)
	}

	// Optional: Adjust pool settings
	// poolConfig.MaxConns = 10
	// poolConfig.MinConns = 2

	// 3. Create the connection pool
	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)

	if err != nil {
		panic(err)
	}

	defer dbPool.Close()

	// 4. Verify the connection
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}
	
	// 1. Intercept CLI migration commands before starting Fyne
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		command := "up"
		var args []string
		if len(os.Args) > 2 {
			command = os.Args[2]
			args = os.Args[3:]
		}

		if err := runMigration(ctx, dbPool, command, args...); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		return // Exit without opening the Fyne window
	}
	*/
	// ----------------------------------------
	// Camera
	// ----------------------------------------
	cameraTypeEnvRow := os.Getenv("CAMERA_TYPE")
	cameraTypeEnv, err := strconv.Atoi(cameraTypeEnvRow)

	if err != nil {
		log.Fatalf("CAMERA_TYPE must be a number, got %q: %w", cameraTypeEnvRow, err)
	}

	cameraType, cameraTypeError := camera.CameraTypeFromInt(cameraTypeEnv)

	if cameraTypeError != nil {
		// Handle invalid config value
		log.Fatalf("Invalid camera type: %v", cameraTypeError)
	}

	log.Println("application run with camera: ", cameraType.String())

	cameraStrategy := camera.NewCameraStrategy()
	cam, err := cameraStrategy.Create(
		camera.CameraConfig{
			Type:        cameraType,
			DevicePath:  cameraDevice,
			Width:       recordingWidth,//cameraWidth,
			Height:      recordingHeight,//cameraHeight,
			FPS:         cameraFPS,
		},
	)

	/*
	cam, err := camera.NewV4LCamera(
		cameraDevice,
		cameraWidth,
		cameraHeight,
		cameraFPS,
	)*/

	//TODO for SCI camera instead NewV4LCamera need to use NewCSICamera
	/*
	cam, err := camera.NewCSICamera(
		cameraWidth,
		cameraHeight,
		cameraFPS,
	)*/

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

	videoPreview := preview.NewFynePreview()

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
	cameraWorkerStrategy := worker.NewCameraFrameWorkerStrategy()

	////Modify for usb cam | Clean for sci
	cameraWorker, cameraWorkerError := cameraWorkerStrategy.Create(
		worker.CameraFrameWorkerConfig{
			WorkerType: cameraType,
			Cam: cam, 
			Detector: detector,
		},
	)
	if cameraWorkerError!= nil {
		log.Fatal(cameraWorkerError)
	}

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
	calcelCtx, cancelCtxFunc := context.WithCancel(
		ctx,
	)

	//avoid context leaks
	defer cancelCtxFunc()

	uiService := ui.NewUIService(
		calcelCtx,
		cancelCtxFunc,
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

//TODO move to separate service
func runMigration(
	ctx context.Context,
	pool *pgxpool.Pool,
	command string,
	args ...string,
) error {
	// Wrap pgxpool.Pool as a standard *sql.DB for Goose
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.RunContext(
		ctx, 
		command, 
		sqlDB, 
		"migrations", 
		args...
	)
}
