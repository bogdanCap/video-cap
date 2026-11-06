package recognition

import (
    "context"
    "image"
)

type FaceRecognizer interface {
    Embed(
        ctx context.Context,
        img image.Image,
    ) ([]float32, error)
}
