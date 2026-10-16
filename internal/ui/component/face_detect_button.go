package component

import (
	"fyne.io/fyne/v2/widget"
)

type FaceDetectButton struct {
    // UI
	button      *widget.Button
}


func NewFaceDetectButton() *FaceDetectButton {
	button :=  &FaceDetectButton{}
	button.button = &widget.Button{
		Text: "face detect",
	}

	return button
	
}

func (d *FaceDetectButton) SetOnTapped(onTapped func()) {
	d.button.OnTapped = onTapped
}

func (d *FaceDetectButton) GetButton() *widget.Button {
	return d.button
}

