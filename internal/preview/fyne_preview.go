package preview

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type FynePreview struct {
	image *canvas.Image
	container *fyne.Container
}

func NewFynePreview() *FynePreview {
	//Initialize UI elements
	img := canvas.NewImageFromImage(
		image.NewRGBA(image.Rect(0, 0, 1, 1)),
	)

	img.FillMode = canvas.ImageFillContain

	content := container.NewWithoutLayout(img)

	return &FynePreview{
		image:     img,
		container: content,
	}
}

func (p *FynePreview) Widget() fyne.CanvasObject {
	return p.container
}

func (p *FynePreview) ShowFrame(
	frame []byte,
) error {

	img, err := jpeg.Decode(
		bytes.NewReader(frame),
	)

	if err != nil {
		return fmt.Errorf(
			"decode JPEG frame: %w",
			err,
		)
	}

	// Fyne UI update.
	fyne.Do(func() {
		p.image.Image = img

		//TODO check fithout this line
		p.image.Resize(
			p.container.Size(),
		)

		p.image.Refresh()
	})

	return nil
}
