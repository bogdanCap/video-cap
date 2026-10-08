package component

import (
	"fyne.io/fyne/v2/widget"
)

type StartButton struct {
    // UI
	button      *widget.Button
}

func NewStartButton(
	/*
    recorder recording.Recorder,
    cam camera.Camera,
    cameraWorker *worker.CameraWorker,
    frameProcessingWorker *worker.FrameProcessingWorker,
	onStopWindow func(),*/
	//videoControlService *service.VideoControlService,
) *StartButton {
	startButton :=  &StartButton{}
	
	startButton.button = &widget.Button{
		Text: "start",
	}

	return startButton
	
}

func (d *StartButton) SetOnTapped(onTapped func()) {
	d.button.OnTapped = onTapped
}

func (d *StartButton) Hide() {
	d.button.Hide()
}

func (d *StartButton) Show() {
	d.button.Show()
}

func (d *StartButton) GetButton() *widget.Button {
	return d.button
}


//func (startButton *StartButton) start() {
//}
