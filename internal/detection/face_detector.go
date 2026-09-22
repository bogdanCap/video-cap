package detection

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"os"

	"github.com/esimov/pigo/core"
	"golang.org/x/image/draw"
)

/*
type Face struct {
	X      int
	Y      int
	Width  int
	Height int
}*/

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

func (fd *FaceDetector) Detect(frame []byte) ([]Face, error) {

	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}

	//return origin image resolution
	origBounds := img.Bounds()
	origWidth := origBounds.Dx()
	origHeight := origBounds.Dy()

	targetWidth := 320
	targetHeight := (origHeight * targetWidth) / origWidth

	smallImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.NearestNeighbor.Scale(smallImg, smallImg.Bounds(), img, origBounds, draw.Src, nil)

	pixels := fd.rgbaToGrayscale(smallImg, targetWidth, targetHeight)

	cParams := pigo.CascadeParams{
		MinSize:     30,
		MaxSize:     600,
		ScaleFactor: 1.1,
		ShiftFactor: 0.1,
		ImageParams: pigo.ImageParams{
			Pixels: pixels,
			Rows:   targetHeight,
			Cols:   targetWidth,
			Dim:    targetWidth,
		},
	}

	// 0.0 means no rotation angle (standard upright face detection)
	angle := 0.0 
	
	// Pass the correct structure and the float64 angle parameter
	faces := fd.classifier.RunCascade(cParams, angle)

	faces = fd.classifier.ClusterDetections(faces, 0.2)

	scaleX := float64(origWidth) / float64(targetWidth)
	scaleY := float64(origHeight) / float64(targetHeight)

	minFaceWidth := 150 
	var detectedFaces []Face

	for _, face := range faces {
		if face.Q > 5.0 {
			radius := face.Scale / 2
			
			smallX := face.Col - radius
			smallY := face.Row - radius
			smallW := face.Scale
			smallH := face.Scale

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
		}
	}

	return detectedFaces, nil
}

// DrawFaces takes the original frame and a slice of Face structs
func (fd *FaceDetector) DrawFaces(frame []byte, faces []Face) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	for _, f := range faces {
		// Clean and scannable interface: passing the entire Face object
		fd.drawBox(rgbaImg, f, green)
	}

	var outBuf bytes.Buffer
	err = jpeg.Encode(&outBuf, rgbaImg, &jpeg.Options{Quality: 90})
	if err != nil {
		return nil, err
	}

	return outBuf.Bytes(), nil
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