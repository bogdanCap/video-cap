package facemove

import (
	//"math"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	//"github.com/bogdanCap/video-cap/internal/detection"
)

type MotionTracker struct {
	previousFrame image.Image

	// Percentage of changed pixels required
	// to detect movement.
	// Example: 2.0 = 2% of pixels changed.
	moveThreshold float64
}



func NewMotionTracker(moveThreshold float64) *MotionTracker {
	return &MotionTracker{
		moveThreshold: moveThreshold,
	}
}

func (mt *MotionTracker) TrackMovement(frame []byte) (bool, error) {
	// Decode JPEG frame.
	currentFrame, err := jpeg.Decode(bytes.NewReader(frame))
	if err != nil {
		return false, fmt.Errorf("decode frame: %w", err)
	}

	// First frame.
	if mt.previousFrame == nil {
		mt.previousFrame = currentFrame

		return false, nil
	}

	currentBounds := currentFrame.Bounds()
	previousBounds := mt.previousFrame.Bounds()

	// Make sure both frames have the same dimensions.
	if currentBounds != previousBounds {
		mt.previousFrame = currentFrame

		return false, fmt.Errorf(
			"frame dimensions changed: current=%v previous=%v",
			currentBounds,
			previousBounds,
		)
	}

	changedPixels := 0
	totalPixels := currentBounds.Dx() * currentBounds.Dy()

	const pixelDifferenceThreshold = 50 * 257

	for y := currentBounds.Min.Y; y < currentBounds.Max.Y; y++ {
		for x := currentBounds.Min.X; x < currentBounds.Max.X; x++ {
			currentR, currentG, currentB, _ :=
				currentFrame.At(x, y).RGBA()

			previousR, previousG, previousB, _ :=
				mt.previousFrame.At(x, y).RGBA()

			rDiff := absInt(int(currentR) - int(previousR))
			gDiff := absInt(int(currentG) - int(previousG))
			bDiff := absInt(int(currentB) - int(previousB))

			if rDiff > pixelDifferenceThreshold ||
				gDiff > pixelDifferenceThreshold ||
				bDiff > pixelDifferenceThreshold {

				changedPixels++
			}
		}
	}

	changedPercentage :=
		float64(changedPixels) / float64(totalPixels) * 100

	// Save current frame for the next comparison.
	mt.previousFrame = currentFrame

	return changedPercentage >= mt.moveThreshold, nil
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}

/*
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
	*/