init:
    go mod init video
    go get fyne.io/fyne/v2@v2.7.0  (this version works on ubuntu 20.04)
    go get github.com/korandiz/v4l@v1.1.0
    
    go get gocv.io/x/gocv@v0.22.0

    go mod edit -go=1.26.0
    got mod tidy
    sudo apt install libgl1-mesa-dev xorg-dev
    sudo apt install ffmpeg
    

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