package component

import (
	"fyne.io/fyne/v2/widget"
)

type MoveDetectButton struct {
    // UI
	button      *widget.Button
}

func NewMoveDetectButton() *MoveDetectButton {
	button :=  &MoveDetectButton{}
	button.button = &widget.Button{
		Text: "move detect",
	}

	return button
	
}

func (d *MoveDetectButton) SetOnTapped(onTapped func()) {
	d.button.OnTapped = onTapped
}

func (d *MoveDetectButton) GetButton() *widget.Button {
	return d.button
}

func (d *MoveDetectButton) SetButton(button *widget.Button) {
	d.button = button
}

