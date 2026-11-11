init:
    go mod init video
    go get fyne.io/fyne/v2@v2.7.0  (this version works on ubuntu 20.04)
    go get github.com/korandiz/v4l@v1.1.0
    
    go get gocv.io/x/gocv@v0.22.0  (we use pigo istead - old version for v4l_camera now do not need)
    go get github.com/esimov/pigo/core

    go get github.com/pgvector/pgvector-go
    go get github.com/yalue/onnxruntime_go  (https://netron.app/?utm_source=chatgpt.com)
    go get golang.org/x/image/draw
    go get github.com/jackc/pgx/v5/pgxpool
    go get github.com/joho/godotenv
    
    go mod edit -go=1.27.0
    got mod tidy
    sudo apt install libgl1-mesa-dev xorg-dev
    sudo apt install ffmpeg
    sudo snap install goose

    go get github.com/pressly/goose/v3
        go install github.com/pressly/goose/v3/cmd/goose@latest
        +
        export PATH=$PATH:$(go env GOPATH)/bin
    go get github.com/jackc/pgx/v5/stdlib

Migration:
    goose -dir migrations create create_users_table sql
    goose -dir migrations postgres "postgres://db_user:secure_password_123@localhost:5432/app_record?sslmode=disable" up

    DATABASE_URL="postgres://user:password@localhost:5432/mydb?sslmode=disable" go run . migrate up
    DATABASE_URL="postgres://user:password@localhost:5432/mydb?sslmode=disable" ./my-fyne-app migrate up

db:
    docker exec -it app_postgres_db pg_isready -U db_user -d app_record

    docker compose down -v (down with volums)

for clone project:
    go mod download
    go mod tidy

check usb devide:
    v4l2-ctl --list-devices

Diagnostic:
    # - this need to check version of the tools
    go version
    go list -m fyne.io/fyne/v2
    echo $DISPLAY

    if echo $DISPLAY return x11, add need to run like this:
        GDK_BACKEND=x11 go run .

run:
    go run main.go
    go run .

    go build -o record
    ./record

    for virtual camera batton and logic:
        sudo modprobe v4l2loopback devices=1 video_nr=2 card_label="Go Virtual Camera" exclusive_caps=1

    #for docker
    chmod +x boot.go_sh
    docker-compose up -d
    ./boot.go_sh

    with app db:
        go build -o record
        ./boot.go_sh

    docker:
        docker-compose up --build
        docker-compose up -d
        docker compose down
        docker compose down -v (stop and delete volum)


Debug video:
    video fpv:
        ffprobe -v error \
        -show_entries format=duration \
        -show_entries stream=avg_frame_rate,r_frame_rate \
        -of default=noprint_wrappers=1 \
        video.mp4

Debug delay:
    start := time.Now()
    log.Printf(
		"DrawFaces took: %v",
		time.Since(start),
	)



camera.Camera          → interface ✅
detection.Detector     → interface ✅
recording.Recorder     → interface ✅

CameraWorker           → concrete struct ✅
FrameProcessingWorker  → concrete struct ✅



USB Camera
    ↓
/dev/video0
    ↓
MJPEG 2560x1440 @ 30 FPS
    ↓
Go V4L2
    ↓
┌───────────────┬─────────────────┐
│               │                 │
Fyne preview    FFmpeg recording
│               │
image.Decode    H.264 → MP4



now working this chematic:

USB Camera
    │
    │ MJPEG
    │ 2560x1440
    │ 30 FPS
    ▼
v4l.Capture()
    │
    ├───────────────► JPEG decode ─────► Fyne preview
    │
    │
    └───────────────► FFmpeg stdin
                            │
                            ▼
                         libx264
                            │
                            ▼
                   recordings/video_*.mp4



Video goroutines schematic:

                USB Camera
                      |
                      v
                 Read Frame
                      |
                      v
             FrameProcessingWorker
                      |
             +--------+--------+
             |        |        |
             v        v        v
          Flip      Detect    Draw
          Frame      Face     Face
             |        |        |
             +--------+--------+
                      |
                      v
              Processed Frame
                      |
               +------+------+
               |             |
               v             v
         Preview Channel  Recording Channel
               |             |
               v             v
         Preview Worker  Recording Worker
               |             |
               v             v
           Fyne UI       FFmpeg/MP4



Virtual camera logic:
    sudo apt install v4l2loopback-dkms v4l2loopback-utils
    sudo modprobe v4l2loopback devices=1 video_nr=2 card_label="Go Virtual Camera" exclusive_caps=1

    v4l2-ctl --list-devices - check if device exists

    this create virtual device - /dev/video2
    and now we can run logic for virtual device stream


USB Camera
    |
    v
CameraWorker
    |
    v
FrameProcessingWorker.PushJob()
    |
    +----> previewChan ------> previewWorker
    |
    +----> recordingChan ----> recordingWorker
    |
    +----> faceMotionChan ---> motionTrackerWorker
    |
    +----> virtualCameraChan -> VirtualCameraWorker
                                    |
                                    | Only when enabled
                                    v
                              FFmpeg stdin
                                    |
                                    v
                               /dev/video2
                                    |
                                    v
                              Google Meet