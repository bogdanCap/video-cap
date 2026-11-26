package preview

import (
	//"bytes"
	//"fmt"
	"image"
	"image/color"
	"log"

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
	//return m.At(x, y)
	
	bounds := m.Image.Bounds()

    mirroredX := bounds.Min.X + bounds.Max.X - 1 - x

    return m.Image.At(mirroredX, y)
}

type FynePreview struct {
	image *canvas.Image
	container *fyne.Container
	//buf    *image.RGBA
	bufs      [2]*image.RGBA
	bufIdx    int
}

const (
	previewWidth = 480//640
	previewHeight = 360
)

func NewFynePreview(
	//width int,
	//height int,
) *FynePreview {
	/*
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
	}*/

	buf0 := image.NewRGBA(image.Rect(0, 0, previewWidth, previewHeight))
	buf1 := image.NewRGBA(image.Rect(0, 0, previewWidth, previewHeight))

	img := canvas.NewImageFromImage(buf0)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScalePixels
	img.SetMinSize(fyne.NewSize(float32(previewWidth), float32(previewHeight)))

	// container.NewStack automatically stretches img to match the container size
	content := container.NewStack(img)

	return &FynePreview{
		image:     img,
		container: content,
		bufs:      [2]*image.RGBA{buf0, buf1},
	}
}

func (p *FynePreview) Widget() fyne.CanvasObject {
	return p.container
}

func (p *FynePreview) ShowFrame(
	img image.Image/*[]byte*/,
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

	

	// 2. Downscale 1280x720 -> 640x360 directly into the reused buffer
	// Using draw.Src skips alpha blending for maximum speed
	/*
	draw.NearestNeighbor.Scale(p.buf, p.buf.Bounds(), frame, frame.Bounds(), draw.Src, nil)


	// Fyne UI update.
	fyne.Do(func() {
		p.image.Image = p.buf//img
		//p.image.Image = flippedRGBA

		/*
		if targetSize := p.container.Size(); p.image.Size() != targetSize {
			p.image.Resize(targetSize)
		}*/
	/*
		p.image.Resize(
			p.container.Size(),
		)

		p.image.Refresh()
	})*/



	if img == nil {
		log.Println("img is nil and preview do not works")

		return nil
	}

	// Pick the back buffer that Fyne is NOT currently rendering
	dst := p.bufs[p.bufIdx]
	p.bufIdx = (p.bufIdx + 1) % 2

	// Fast downscale into the back buffer
	draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	/* this need if preview resolution with incorrect ratio
	draw.BiLinear.Scale(
		dst, 
		dst.Bounds(),
		 img, 
		 img.Bounds(),
		draw.Src, 
		nil,
	)*/

	// Swap buffer pointer on the Fyne UI thread
	fyne.Do(func() {
		p.image.Image = dst

		p.image.Resize(
			p.container.Size(),
		)

		p.image.Refresh()
	})

	return nil
}
