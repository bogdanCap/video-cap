package component

import (
	"fyne.io/fyne/v2/widget"
)

type StopButton struct {
    // UI
	button      *widget.Button
}

func NewStopButton() *StopButton {
	stopButton :=  &StopButton{}
	stopButton.button = &widget.Button{
		Text: "stop",
	}

	return stopButton
	
}

func (d *StopButton) SetOnTapped(onTapped func()) {
	d.button.OnTapped = onTapped
}

func (d *StopButton) Hide() {
	d.button.Hide()
}

func (d *StopButton) Show() {
	d.button.Show()
}

func (d *StopButton) GetButton() *widget.Button {
	return d.button
}
