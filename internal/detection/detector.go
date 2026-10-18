package detection

import (
	"image"
)

type Detector interface {
	Detect(img image.Image/*, origWidth int, origHeight int, targetWidth int, targetHeight int*/) ([]Face, error)
	DrawFaces(
		img image.Image,
		faces []Face,
	) (image.Image, error)
	/*ScaleFaces(
		faces []Face,
		smallWidth int,
		smallHeight int,
		originalWidth int,
		originalHeight int,
	) []Face*/
}
