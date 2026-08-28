package detection

type Face struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Detector interface {
	Detect(frame []byte) ([]Face, error)
	DrawFaces(
		frame []byte,
		faces []Face,
	) ([]byte, error)
}
