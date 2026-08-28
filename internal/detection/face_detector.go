package detection

import (
	"fmt"

	"gocv.io/x/gocv"
	"image"
	"image/color"
)

type FaceDetector struct {
	classifier gocv.CascadeClassifier
}

func NewFaceDetector(cascadePath string) (*FaceDetector, error) {
	classifier := gocv.NewCascadeClassifier()

	if !classifier.Load(cascadePath) {
		classifier.Close()

		return nil, fmt.Errorf(
			"failed to load face cascade: %s",
			cascadePath,
		)
	}

	return &FaceDetector{
		classifier: classifier,
	}, nil
}

func (d *FaceDetector) Detect(frame []byte) ([]Face, error) {

	// JPEG []byte -> OpenCV Mat.
	mat, err := gocv.IMDecode(
		frame,
		gocv.IMReadColor,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode frame: %w",
			err,
		)
	}

	defer mat.Close()

	if mat.Empty() {
		return nil, fmt.Errorf(
			"decoded frame is empty",
		)
	}

	// Convert BGR -> grayscale.
	gray := gocv.NewMat()
	defer gray.Close()

	gocv.CvtColor(
		mat,
		&gray,
		gocv.ColorBGRToGray,
	)

	// Improve contrast.
	gocv.EqualizeHist(
		gray,
		&gray,
	)

	// Actual face detection.
	rects := d.classifier.DetectMultiScale(
		gray,
	)

	/* for multiple faces
	faces := make(
		[]Face,
		0,
		len(rects),
	)

	for _, rect := range rects {
		faces = append(
			faces,
			Face{
				X:      rect.Min.X,
				Y:      rect.Min.Y,
				Width:  rect.Dx(),
				Height: rect.Dy(),
			},
		)
	}

	return faces, nil*/

	//for single face
	if len(rects) == 0 {
		return nil, nil
	}

	// Only one face.
	rect := rects[0]

	return []Face{
		{
			X:      rect.Min.X,
			Y:      rect.Min.Y,
			Width:  rect.Dx(),
			Height: rect.Dy(),
		},
	}, nil
}

func (d *FaceDetector) DrawFaces(
	frame []byte,
	faces []Face,
) ([]byte, error) {

	mat, err := gocv.IMDecode(
		frame,
		gocv.IMReadColor,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode frame: %w",
			err,
		)
	}

	defer mat.Close()

	// Draw face rectangles.
	for _, face := range faces {

		rect := image.Rect(
			face.X,
			face.Y,
			face.X+face.Width,
			face.Y+face.Height,
		)

		gocv.Rectangle(
			&mat,
			rect,
			color.RGBA{
				R: 255,
				G: 0,
				B: 0,
				A: 255,
			},
			4,
		)
	}

	// Encode Mat back to JPEG.
	result, err := gocv.IMEncode(
		gocv.JPEGFileExt,
		mat,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"encode frame: %w",
			err,
		)
	}

	// result is already []byte.
	return result, nil
}

func (d *FaceDetector) Close() {
	d.classifier.Close()
}
