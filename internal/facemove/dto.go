package facemove

import (
	"github.com/bogdanCap/video-cap/internal/detection"
)

type MotionResult struct {
	IsMoving bool
	Face     detection.Face
	Frame    []byte
}