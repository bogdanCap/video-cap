package detection

type Detector interface {
	Detect(frame []byte) ([]Face, error)
	DrawFaces(
		frame []byte,
		faces []Face,
	) ([]byte, error)
}
