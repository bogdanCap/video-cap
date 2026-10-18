package detection

import (
	//"bytes"
	"image"
	"image/color"
	//"image/jpeg"
	//_ "image/jpeg"
	"os"
	"log"

	"github.com/esimov/pigo/core"
	"golang.org/x/image/draw"
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

func (fd *FaceDetector) Detect(img image.Image/*, origWidth int, origHeight int, targetWidth int, targetHeight int*/) ([]Face, error) {

	/* this logic reduce image - i moved it to another package
	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}*/

	//return origin image resolution
	/*
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	targetWidth := 320
	targetHeight := (origHeight * targetWidth) / origWidth

	smallImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.NearestNeighbor.Scale(
		smallImg, 
		smallImg.Bounds(), 
		img, 
		origBounds, 
		draw.Src, 
		nil,
	)

	pixels := fd.rgbaToGrayscale(smallImg, targetWidth, targetHeight)
	*/

	//this do not need if i reduce image size in logic camera.ReduceImage(imgFrame, 640)
	//i get this from method parameter
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	pixels := fd.rgbaToGrayscale(img, origWidth, origHeight)

	cParams := pigo.CascadeParams{
		MinSize:     30,
		MaxSize:     600,
		ScaleFactor: 1.1,
		ShiftFactor: 0.1,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   origHeight,
			Cols:   origWidth,
			Dim:    origWidth,
		},
	}

	// 0.0 means no rotation angle (standard upright face detection)
	angle := 0.0 
	
	// Pass the correct structure and the float64 angle parameter
	faces := fd.classifier.RunCascade(cParams, angle)

	faces = fd.classifier.ClusterDetections(faces, 0.2)

	/* this need if i reduce image size in logic camera.ReduceImage(imgFrame, 640)
	scaleX := float64(origWidth) / float64(targetWidth)
	scaleY := float64(origHeight) / float64(targetHeight)
	*/
	scaleX := 1.0 //float64(origWidth) / float64(targetWidth)
	scaleY := 1.0 //float64(origHeight) / float64(targetHeight)

	//faceSquareScale := 1.0 // if reduce this value need to reduce variable minFaceWidth to show this
	minFaceWidth := 150 
	var detectedFaces []Face

	for _, face := range faces {
		//logic with rectangle
		/*
		if face.Q > 5.0 {
			radius := face.Scale / 2
			
			smallX := face.Col - radius
			smallY := face.Row - radius
			// Increase the square diameter.
    		smallW := int(float64(face.Scale) * faceSquareScale)
    		smallH := int(float64(face.Scale) * faceSquareScale)
			//smallW := face.Scale
			//smallH := face.Scale

			// Map coordinates back up to full-scale size
			fullWidth := int(float64(smallW) * scaleX)
			//fullHeight := int(float64(smallH) * scaleY)

			// 🚫 SIZE FILTER: Only return the face if it meets your minimum size requirement
			if fullWidth < minFaceWidth {
				continue // Skip this small face entirely
			}

			detectedFaces = append(detectedFaces, Face{
				X:      int(float64(smallX) * scaleX),
				Y:      int(float64(smallY) * scaleY),
				Width:  int(float64(smallW) * scaleX),
				Height: int(float64(smallH) * scaleY),
			})

			//only need 1 face object
			break
		}*/

		//logic with face detection option 2
		if face.Q <= 5.0 {
			continue
		}

		fullWidth := int(float64(face.Scale) * scaleX)

		if fullWidth < minFaceWidth {
			continue
		}

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
func (fd *FaceDetector) DrawFaces(img image.Image, faces []Face) (image.Image/*[]byte*/, error) {
	/* logic to draw rectangle

	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	for _, f := range faces {
		// Clean and scannable interface: passing the entire Face object
		fd.drawBox(rgbaImg, f, green)
	}

	return rgbaImg, nil
	*/

	bounds := img.Bounds()

	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	crossSize := 10

	for _, f := range faces {
		// Face center in image coordinates.
		centerX := bounds.Min.X + f.X + f.Width/2
		centerY := bounds.Min.Y + f.Y + f.Height/2

		log.Printf(
			"Drawing cross: center=(%d,%d), image bounds=%v",
			centerX, centerY, bounds,
		)

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

func (fd *FaceDetector) drawLine(
	img *image.RGBA,
	x1, y1, x2, y2 int,
	lineWidth int,
	c color.RGBA,
) {
	bounds := img.Bounds()
	halfWidth := lineWidth / 2

	// Horizontal line.
	for x := x1; x <= x2; x++ {
		for y := y1 - halfWidth; y <= y1+halfWidth; y++ {
			if image.Pt(x, y).In(bounds) {
				img.SetRGBA(x, y, c)
			}
		}
	}

	// Vertical line.
	for y := y1; y <= y2; y++ {
		for x := x1 - halfWidth; x <= x1+halfWidth; x++ {
			if image.Pt(x, y).In(bounds) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func ScaleFaces(
	faces []Face,
	smallWidth int,
	smallHeight int,
	originalWidth int,
	originalHeight int,
	mirrored bool,
) []Face {
	scaleX := float64(originalWidth) / float64(smallWidth)
	scaleY := float64(originalHeight) / float64(smallHeight)

	scaledFaces := make([]Face, 0, len(faces))

	for _, face := range faces {
		x := int(float64(face.X) * scaleX)
		y := int(float64(face.Y) * scaleY)
		width := int(float64(face.Width) * scaleX)
		height := int(float64(face.Height) * scaleY)

		if mirrored {
			x = originalWidth - x - width
		}

		scaledFaces = append(scaledFaces, Face{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		})
	}

	return scaledFaces
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

// drawBox now cleanly consumes the structured Face model directly
func (fd *FaceDetector) drawBox(img *image.RGBA, f Face, col color.Color) {
	bounds := img.Bounds()

	// Draw top and bottom horizontal borders
	for i := f.X; i < f.X+f.Width; i++ {
		if i >= bounds.Min.X && i < bounds.Max.X {
			if f.Y >= bounds.Min.Y && f.Y < bounds.Max.Y {
				img.Set(i, f.Y, col)
			}
			if f.Y+f.Height >= bounds.Min.Y && f.Y+f.Height < bounds.Max.Y {
				img.Set(i, f.Y+f.Height, col)
			}
		}
	}
	
	// Draw left and right vertical borders
	for j := f.Y; j < f.Y+f.Height; j++ {
		if j >= bounds.Min.Y && j < bounds.Max.Y {
			if f.X >= bounds.Min.X && f.X < bounds.Max.X {
				img.Set(f.X, j, col)
			}
			if f.X+f.Width >= bounds.Min.X && f.X+f.Width < bounds.Max.X {
				img.Set(f.X+f.Width, j, col)
			}
		}
	}


	/* if need change line thickness
	lineHeight := 5  // Controls thickness of TOP and BOTTOM horizontal bars
	lineWidth  := 2  // Controls thickness of LEFT and RIGHT vertical bars

	// 1. Draw Top and Bottom Horizontal Bars
	for h := 0; h < lineHeight; h++ { // 👈 Uses your custom line height
		topY := f.Y + h
		bottomY := f.Y + f.Height - h

		for x := f.X; x < f.X+f.Width; x++ {
			fd.setPixelSafe(img, x, topY, bounds, col)
			fd.setPixelSafe(img, x, bottomY, bounds, col)
		}
	}

	// 2. Draw Left and Right Vertical Bars
	for w := 0; w < lineWidth; w++ { // 👈 Uses your custom line width
		leftX := f.X + w
		rightX := f.X + f.Width - w

		for y := f.Y; y < f.Y+f.Height; y++ {
			fd.setPixelSafe(img, leftX, y, bounds, col)
			fd.setPixelSafe(img, rightX, y, bounds, col)
		}
	}*/
}

/*
// Helper to safely paint pixels only if they fall inside the valid image canvas bounds
func (fd *FaceDetector) setPixelSafe(img *image.RGBA, x, y int, bounds image.Rectangle, col color.Color) {
	if x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y {
		img.Set(x, y, col)
	}
}*/