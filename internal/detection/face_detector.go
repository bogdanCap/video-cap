package detection

import (
	"image"
	"image/color"
	"os"
	"github.com/esimov/pigo/core"
	"golang.org/x/image/draw"
	"github.com/bogdanCap/video-cap/internal/camera"
)

type Face struct {
	X      int
	Y      int
	Width  int
	Height int
}

type FaceDetector struct {
	classifier *pigo.Pigo
}

func NewFaceDetector(cascadePath string) (*FaceDetector, error) {
	cascadeFile, err := os.ReadFile(cascadePath)
	if err != nil {
		return nil, err
	}

	p := pigo.NewPigo()
	classifier, err := p.Unpack(cascadeFile)
	if err != nil {
		return nil, err
	}

	return &FaceDetector{classifier: classifier}, nil
}

func (fd *FaceDetector) Detect(img image.Image) ([]Face, error) {
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	// resize image logic
	// Reduce resolution for faster face detection.
    targetWidth := 320
    //targetHeight := 240

    smallImg := camera.ReduceImage(img, targetWidth)
	targetHeight := smallImg.Bounds().Dy()

	
	pixels := fd.rgbaToGrayscale(
		/*
		img, 
		origWidth, 
		origHeight,
		*/
		smallImg,
		targetWidth,
		targetHeight,
	)

	cParams := pigo.CascadeParams{
		//Ignore any face smaller than
		MinSize:     30,
		//Ignore any face larger than
		MaxSize:     targetHeight/*240*//*600*/,
		ScaleFactor: 1.2,
		ShiftFactor: 0.2,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   targetHeight,
            Cols:   targetWidth,
            Dim:    targetWidth,
			//Rows:   origHeight,
			//Cols:   origWidth,
			//Dim:    origWidth,
		},
	}

	// 0.0 means no rotation angle (standard upright face detection)
	angle := 0.0 
	
	// Pass the correct structure and the float64 angle parameter
	faces := fd.classifier.RunCascade(cParams, angle)

	// iouThreshold - decimal fraction between 0.0 (0% overlap) and 1.0 (100% perfect overlap)
	// 0.0: The two boxes with face do not touch at all
	// 1.0: The two boxes match perfectly, corner for corner.
	faces = fd.classifier.ClusterDetections(faces, 0.2)

	// this need if i reduce image size in logic camera.ReduceImage(imgFrame, 640)
	scaleX := float64(origWidth) / float64(targetWidth)
	scaleY := float64(origHeight) / float64(targetHeight)
	
	//scaleX := 1.0 //float64(origWidth) / float64(targetWidth)
	//scaleY := 1.0 //float64(origHeight) / float64(targetHeight)

	//faceSquareScale := 1.0 // if reduce this value need to reduce variable minFaceWidth to show this
	minFaceWidth := 150 
	var detectedFaces []Face

	for _, face := range faces {
		//logic with rectangle

		//found a potential face - from 0.0 to 10.0 - The code becomes much stricter
		if face.Q <= 5.0 {
			continue
		}

		fullWidth := int(float64(face.Scale) * scaleX)

		if fullWidth < minFaceWidth {
			continue
		}

		//The radius calculation stays the same. Only the scale factors change.
		radius := face.Scale / 2

		smallX := face.Col - radius
		smallY := face.Row - radius

		detectedFaces = append(detectedFaces, Face{
			X:      int(float64(smallX) * scaleX),
			Y:      int(float64(smallY) * scaleY),
			Width:  int(float64(face.Scale) * scaleX),
			Height: int(float64(face.Scale) * scaleY),
		})

		break // Only one face.
	}

	return detectedFaces, nil
}

// DrawFaces takes the original frame and a slice of Face structs
func (fd *FaceDetector) DrawFaces(img image.Image, faces []Face) (image.Image, error) {
	bounds := img.Bounds()

	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	crossSize := 10

	for _, f := range faces {
		// Face center in image coordinates.
		centerX := bounds.Min.X + f.X + f.Width/2
		centerY := bounds.Min.Y + f.Y + f.Height/2

		for i := -crossSize; i <= crossSize; i++ {
			// Horizontal line.
			x := centerX + i
			if image.Pt(x, centerY).In(bounds) {
				rgbaImg.SetRGBA(x, centerY, green)
			}

			// Vertical line.
			y := centerY + i
			if image.Pt(centerX, y).In(bounds) {
				rgbaImg.SetRGBA(centerX, y, green)
			}
		}
	}

	return rgbaImg, nil
}

func (fd *FaceDetector) rgbaToGrayscale(img image.Image, width, height int) []uint8 {
	gray := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray[y*width+x] = uint8((0.299 * float64(r>>8)) + (0.587 * float64(g>>8)) + (0.114 * float64(b>>8)))
		}
	}
	return gray
}
