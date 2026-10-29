package camera

import (
	//"bytes"
	//"fmt"
	"image"
	//"image/jpeg"
	//"image/draw"
	"golang.org/x/image/draw"
)

//TODO move to another package
func FlipImage(img image.Image) (image.Image, error) {
	switch src := img.(type) {
	case *image.YCbCr:
		return flipYCbCr(src), nil
	case *image.RGBA:
		return flipRGBA(src), nil
	default:
		// Support other image types by converting to RGBA.
		bounds := img.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

		return flipRGBA(rgba), nil
	}
}

func flipYCbCr(src *image.YCbCr) *image.YCbCr {
	bounds := src.Bounds()
	dst := image.NewYCbCr(bounds, src.SubsampleRatio)

	width := bounds.Dx()
	height := bounds.Dy()

	// Flip luminance plane.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := bounds.Max.X - 1 - x
			srcY := bounds.Min.Y + y
			dstX := bounds.Min.X + x
			dstY := bounds.Min.Y + y

			dst.Y[dst.YOffset(dstX, dstY)] =
				src.Y[src.YOffset(srcX, srcY)]
		}
	}

	// Flip chroma by sampling each source pixel and copying
	// the corresponding Cb/Cr values to the mirrored destination.
	// This works with all YCbCr subsampling ratios.
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.COffset(bounds.Max.X - 1 - (x - bounds.Min.X), y)
			d := dst.COffset(x, y)

			dst.Cb[d] = src.Cb[c]
			dst.Cr[d] = src.Cr[c]
		}
	}

	return dst

	//if need to disabled
	//return src
}

func flipRGBA(src *image.RGBA) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(
				bounds.Min.X+bounds.Max.X-1-x,
				y,
				src.At(x, y),
			)
		}
	}

	return dst
	//if need to disabled
	//return src
}

func ReduceImage(
    originImg image.Image,
    targetWidth int,
) (*image.RGBA) {
    origBounds := originImg.Bounds()
    origWidth := origBounds.Dx()
    origHeight := origBounds.Dy()

    targetHeight := origHeight * targetWidth / origWidth

    smallImg := image.NewRGBA(
        image.Rect(0, 0, targetWidth, targetHeight),
    )

    draw.NearestNeighbor.Scale(
        smallImg,
        smallImg.Bounds(),
        originImg,
        origBounds,
        draw.Src,
        nil,
    )

    return smallImg

	/** option 2 for reduce size
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	scaleX := float64(srcWidth) / float64(targetWidth)
	scaleY := float64(srcHeight) / float64(targetHeight)

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			dst.Set(x, y, img.At(
				srcBounds.Min.X+srcX,
				srcBounds.Min.Y+srcY,
			))
		}
	}

	return dst
	**/
}