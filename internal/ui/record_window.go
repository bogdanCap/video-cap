package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

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

	videoControls := NewVideoControl(
        recorder,
        cam,
        cameraWorker,
        frameProcessingWorker,
		//callback to stop fyne window
		func() {
			myWindow.Close()
		},
    )

    content := container.NewBorder(
        nil,
        videoControls.Buttons(),
        nil,
        nil,
        videoPreview.Widget(),
    )

    myWindow.SetContent(content)

    myWindow.Resize(
        fyne.NewSize(800, 480),
    )

    return &WindowService{
        App:    myApp,
        Window: myWindow,
    }
}

func (s *WindowService) Start() {
	s.Window.ShowAndRun()
}

func (s *WindowService) Close() {
	s.Window.Close()
}
