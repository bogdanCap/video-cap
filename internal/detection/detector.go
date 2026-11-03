package detection

import (
	"image"
)

type Detector interface {
	Detect(img image.Image) ([]Face, error)
	DrawFaces(
		img image.Image,
		faces []Face,
	) (image.Image, error)
}
