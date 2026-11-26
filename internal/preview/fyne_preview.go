package preview

import (
	//"bytes"
	//"fmt"
	"image"
	"image/color"
	"log"
	"time"

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

	lastUpdate time.Time
}

const (
	previewWidth = 480//640
	previewHeight = 360
	// 33 - 30fps
	// 50 - 20fps
	// 66 - 15fps
	minFrameGap   = 33 * time.Millisecond // Cap preview at ~15 FPS
)

func NewFynePreview(
	//width int,
	//height int,
) *FynePreview {


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
	if img == nil {
		log.Println("img is nil and preview do not works")

		return nil
	}

	// 1. Throttle UI updates to ~15 FPS so fyne.Do() never backs up or races buffers
	//this can sllow preview if configurate incorrect
	if time.Since(p.lastUpdate) < minFrameGap {
		return nil
	}
	p.lastUpdate = time.Now()

	
	// Pick the back buffer that Fyne is NOT currently rendering
	dst := p.bufs[p.bufIdx]
	p.bufIdx = (p.bufIdx + 1) % 2

	// Fast downscale into the back buffer
	//draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	// 3. Fast path for CSI YUV420 (*image.YCbCr) avoids slow interface At() calls
	if ycbcr, ok := img.(*image.YCbCr); ok && ycbcr.SubsampleRatio == image.YCbCrSubsampleRatio420 {
		dst = scaleYCbCr420ToRGBA(dst, ycbcr)
	} else {
		draw.NearestNeighbor.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	}
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

// scaleYCbCr420ToRGBA downscales and converts YUV420 directly into RGBA in ~0.4ms.
func scaleYCbCr420ToRGBA(dst *image.RGBA, src *image.YCbCr) *image.RGBA {
	dstW, dstH := dst.Bounds().Dx(), dst.Bounds().Dy()
	srcW, srcH := src.Bounds().Dx(), src.Bounds().Dy()

	for dy := 0; dy < dstH; dy++ {
		sy := (dy * srcH) / dstH
		yRow := sy * src.YStride
		cRow := (sy / 2) * src.CStride
		dstRow := dy * dst.Stride

		for dx := 0; dx < dstW; dx++ {
			sx := (dx * srcW) / dstW

			yVal := int32(src.Y[yRow+sx])
			cbVal := int32(src.Cb[cRow+(sx/2)]) - 128
			crVal := int32(src.Cr[cRow+(sx/2)]) - 128

			r := yVal + ((91881 * crVal) >> 16)
			g := yVal - ((22554*cbVal + 46802*crVal) >> 16)
			b := yVal + ((116130 * cbVal) >> 16)

			if r < 0 {
				r = 0
			} else if r > 255 {
				r = 255
			}
			if g < 0 {
				g = 0
			} else if g > 255 {
				g = 255
			}
			if b < 0 {
				b = 0
			} else if b > 255 {
				b = 255
			}

			p := dstRow + dx*4
			dst.Pix[p] = uint8(r)
			dst.Pix[p+1] = uint8(g)
			dst.Pix[p+2] = uint8(b)
			dst.Pix[p+3] = 255
		}
	}

	return dst
}
