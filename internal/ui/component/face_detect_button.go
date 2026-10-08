package component

import (
	"fyne.io/fyne/v2/widget"
)

type FaceDetectButton struct {
    // UI
	button      *widget.Button
}


func NewFaceDetectButton(
	/*
    recorder recording.Recorder,
    cam camera.Camera,
    cameraWorker *worker.CameraWorker,
    frameProcessingWorker *worker.FrameProcessingWorker,
	onStopWindow func(),*/
	//videoControlService *service.VideoControlService,
) *FaceDetectButton {
	button :=  &FaceDetectButton{}
	button.button = &widget.Button{
		Text: "face detect",
	}

	return button
	
}

func (d *FaceDetectButton) SetOnTapped(onTapped func()) {
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
func (d *FaceDetectButton) GetButton() *widget.Button {
	return d.button
}


//func (startButton *StartButton) start() {
//}
