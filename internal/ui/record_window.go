package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/bogdanCap/video-cap/internal/camera"
	"github.com/bogdanCap/video-cap/internal/preview"
	"github.com/bogdanCap/video-cap/internal/recording"
	"github.com/bogdanCap/video-cap/internal/worker"
	"github.com/bogdanCap/video-cap/internal/service"
	//"github.com/bogdanCap/video-cap/internal/virtualcamera"
)

type WindowService struct {
	//App    fyne.App
	Window fyne.Window
}

func NewUIService(
	recorder recording.Recorder, 
	cam camera.Camera,
	//virtualCam virtualcamera.VirtualCamera,
	virtualCameraWorker *worker.VirtualCameraWorker,
	frameProcessingWorker *worker.FrameProcessingWorker,
	cameraWorker *worker.CameraWorker,
	videoPreview *preview.FynePreview,
) *WindowService {
	myApp := app.New()

	myWindow := myApp.NewWindow(
		"Recorder",
	)

	videoControlService := service.NewVideoControlService(
		recorder,
        cam,
        cameraWorker,
        frameProcessingWorker,
		virtualCameraWorker,
		//callback to close fyne window
		func() {
			myWindow.Close()
		},
	)

    content := container.NewBorder(
        nil,
        videoControlService.Buttons(),
        nil,
        nil,
        videoPreview.Widget(),
    )

    myWindow.SetContent(content)

    myWindow.Resize(
        fyne.NewSize(800, 480),
		//fyne.NewSize(320, 240),
    )

    return &WindowService{
        //App:    myApp,
        Window: myWindow,
    }
}

func (s *WindowService) Start() {
	s.Window.ShowAndRun()
}

func (s *WindowService) Close() {
	s.Window.Close()
}
