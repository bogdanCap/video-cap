package preview

import (
	//"bytes"
	//"fmt"
	"image"
	"image/color"
	//"image/jpeg"
	"fyne.io/fyne/v2/container"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"golang.org/x/image/draw"
)

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

type FynePreview struct {
	image *canvas.Image
	container *fyne.Container
	buf    *image.RGBA
}

const (
	previewWidth = 640
	previewHeight = 360
)

func NewFynePreview(
	//width int,
	//height int,
) *FynePreview {
	buf := image.NewRGBA(
		image.Rect(0, 0, previewWidth, previewHeight),
	)

	//Initialize UI elements
	img := canvas.NewImageFromImage(buf)

	//// Preserves the aspect ratio
	img.FillMode = canvas.ImageFillContain

	// 3. PIXELS stops blurry linear filtering (retains crisp raw video frames)
	img.ScaleMode = canvas.ImageScalePixels

	// Enforce a minimum base interface dimension
	img.SetMinSize(fyne.NewSize(float32(previewWidth), float32(previewHeight)))

	content := container.NewWithoutLayout(img)

	return &FynePreview{
		image:     img,
		container: content,
		buf: buf,
	}
}

func (p *FynePreview) Widget() fyne.CanvasObject {
	return p.container
}

func (p *FynePreview) ShowFrame(
	frame /*[]byte*/ image.Image,
) error {
	//DOTO trhis already done in camera worker and do not need to di it twice
	/*
	img, err := jpeg.Decode(
		bytes.NewReader(frame),
	)

	if err != nil {
		return fmt.Errorf(
			"decode JPEG frame: %w",
			err,
		)
	}*/

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

	// 2. Downscale 1280x720 -> 640x360 directly into the reused buffer
	// Using draw.Src skips alpha blending for maximum speed
	draw.NearestNeighbor.Scale(p.buf, p.buf.Bounds(), frame, frame.Bounds(), draw.Src, nil)


	// Fyne UI update.
	fyne.Do(func() {
		p.image.Image = p.buf//img
		//p.image.Image = flippedRGBA

		/*
		if targetSize := p.container.Size(); p.image.Size() != targetSize {
			p.image.Resize(targetSize)
		}*/

		p.image.Resize(
			p.container.Size(),
		)

		p.image.Refresh()
	})

	return nil
}
