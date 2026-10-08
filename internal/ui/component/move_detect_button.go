package component

import (
	"fyne.io/fyne/v2/widget"
)

type MoveDetectButton struct {
    // UI
	button      *widget.Button
}

func NewMoveDetectButton(
	/*
    recorder recording.Recorder,
    cam camera.Camera,
    cameraWorker *worker.CameraWorker,
    frameProcessingWorker *worker.FrameProcessingWorker,
	onStopWindow func(),*/
	//videoControlService *service.VideoControlService,
) *MoveDetectButton {
	button :=  &MoveDetectButton{}
	button.button = &widget.Button{
		Text: "move detect",
	}

	return button
	
}

func (d *MoveDetectButton) SetOnTapped(onTapped func()) {
	d.button.OnTapped = onTapped
}
/*
func (d *FaceDetectButton) Hide() {
	d.button.Hide()
}

func (d *FaceDetectButton) Show() {
	d.button.Show()
}
*/
func (d *MoveDetectButton) GetButton() *widget.Button {
	return d.button
}

func (d *MoveDetectButton) SetButton(button *widget.Button) {
	d.button = button
}


//func (startButton *StartButton) start() {
//}
