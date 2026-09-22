package facemove

import (
	"math"

	"github.com/bogdanCap/video-cap/internal/detection"
)

type MotionTracker struct {
	lastX           int
	lastY           int
	hasPreviousFace bool
	
	// Configurable settings
	moveThreshold   float64 // Minimum pixel drift to count as movement (e.g., 15.0 pixels)
}

// NewMotionTracker instantiates the coordinate tracking state
func NewMotionTracker(driftThreshold float64) *MotionTracker {
	return &MotionTracker{
		moveThreshold:   driftThreshold,
		hasPreviousFace: false,
	}
}

// TrackMovement checks if the largest face has moved significantly compared to the last frame
func (mt *MotionTracker) TrackMovement(face detection.Face) (bool, detection.Face, error) {
	// Calculate current center point coordinates
	currentCenterX := face.X + (face.Width / 2)
	currentCenterY := face.Y + (face.Height / 2)

	// Baseline state if a face just appeared on camera
	if !mt.hasPreviousFace {
		mt.lastX = currentCenterX
		mt.lastY = currentCenterY
		mt.hasPreviousFace = true
		return false, face, nil
	}

	// Calculate positional shift distance
	deltaX := float64(currentCenterX - mt.lastX)
	deltaY := float64(currentCenterY - mt.lastY)
	distanceMoved := math.Sqrt(deltaX*deltaX + deltaY*deltaY)

	// Save positions for next pass comparison
	mt.lastX = currentCenterX
	mt.lastY = currentCenterY

	// Return true if distance passes your threshold
	if distanceMoved > mt.moveThreshold {
		return true, face, nil
	}

	return false, face, nil
}