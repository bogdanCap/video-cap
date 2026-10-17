package preview

import (
	"bytes"
	"fmt"
	//"time"
	"image"
	"image/color"
	"image/jpeg"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

/* preview rorate image logic - Option one
import (
	"image"
	"image/color"
)

// MirrorImage wraps any image.Image and flips it horizontally.
type MirrorImage struct {
	image.Image
}

// At overrides the pixel coordinate lookup to create a mirror effect.
func (m MirrorImage) At(x, y int) color.Color {
	bounds := m.Bounds()
	// Reverse the X coordinate relative to the image boundaries
	flippedX := bounds.Max.X - 1 - (x - bounds.Min.X)
	return m.Image.At(flippedX, y)
}*/
// MirrorImage wraps any image.Image and flips it horizontally.
type MirrorImage struct {
	image.Image
	cachedBound image.Rectangle
}

// At overrides the pixel coordinate lookup to create a mirror effect.
func (m MirrorImage) At(x, y int) color.Color {
	return m.At(x, y)
	
	//bounds := m.Bounds()
	// Reverse the X coordinate relative to the image boundaries
	//flippedX := bounds.Max.X - 1 - (x - bounds.Min.X)
	
	//flippedX := m.cachedBound.Max.X - 1 - (x - m.cachedBound.Min.X)
	//return m.At(flippedX, y)
}

/* option 2 - need to test what method faster

img, err := jpeg.Decode(bytes.NewReader(frame))

if err != nil {
	return fmt.Errorf("decode JPEG frame: %w", err)
}

if rgbaImg, ok := img.(*image.RGBA); ok {
	bounds := rgbaImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for y := 0; y < height; y++ {
		rowStart := y * rgbaImg.Stride

		for x := 0; x < width/2; x++ {
			// Calculate exact byte positions directly inside Pix array
			leftIdx  := rowStart + (x * 4)
			rightIdx := rowStart + ((width - 1 - x) * 4)

			// Swap pixels directly inside rgbaImg.Pix
			rgbaImg.Pix[leftIdx], rgbaImg.Pix[rightIdx]         = rgbaImg.Pix[rightIdx], rgbaImg.Pix[leftIdx]     // R
			rgbaImg.Pix[leftIdx+1], rgbaImg.Pix[rightIdx+1] = rgbaImg.Pix[rightIdx+1], rgbaImg.Pix[leftIdx+1] // G
			rgbaImg.Pix[leftIdx+2], rgbaImg.Pix[rightIdx+2] = rgbaImg.Pix[rightIdx+2], rgbaImg.Pix[leftIdx+2] // B
			rgbaImg.Pix[leftIdx+3], rgbaImg.Pix[rightIdx+3] = rgbaImg.Pix[rightIdx+3], rgbaImg.Pix[leftIdx+3] // A
		}
	}
	// Explicitly update the main variable so the edit is visible to readers
	img = rgbaImg
}

// 2. Fyne UI update
fyne.Do(func() {
	p.image.Image = img
	p.image.Resize(p.container.Size())
	p.image.Refresh()
})

*/

type FynePreview struct {
	image *canvas.Image
	container *fyne.Container
}

func NewFynePreview(
	width int,
	height int,
) *FynePreview {
	//Initialize UI elements
	img := canvas.NewImageFromImage(
		image.NewRGBA(image.Rect(0, 0, width, height)),
	)

	img.FillMode = canvas.ImageFillContain

	// 3. PIXELS stops blurry linear filtering (retains crisp raw video frames)
	img.ScaleMode = canvas.ImageScalePixels

	// Enforce a minimum base interface dimension
	img.SetMinSize(fyne.NewSize(float32(width), float32(height)))

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

	/*
	start := time.Now()
	// Wrap the image to flip it like a mirror
	bounds := img.Bounds()
	flippedRGBA := image.NewRGBA(bounds)

	// Process the pixels sequentially
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Calculate the opposite source pixel coordinate
			flippedX := bounds.Max.X - 1 - (x - bounds.Min.X)
			
			// Take the pixel from the mirrored position and place it sequentially at (x, y)
			flippedRGBA.Set(x, y, img.At(flippedX, y))
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Image flip processing took: %s\n", elapsed)*/

	// Fyne UI update.
	fyne.Do(func() {
		p.image.Image = img
		//p.image.Image = flippedRGBA

		//TODO check fithout this line
		p.image.Resize(
			p.container.Size(),
		)

		p.image.Refresh()
	})

	return nil
}
